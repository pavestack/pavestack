package apiserver

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// catalogInfoYAML is the subset of the Backstage catalog-info.yaml shape
// Pavestack scaffolds. This is the same file pavestack-portal's
// generate-catalog.mjs reads — see decisions.md for why this Go reader and
// the portal's JS reader intentionally target the identical files instead
// of a shared library: they run in different runtimes and the schema is
// small enough that duplicating the *reader* (not the data) is cheaper than
// a cross-language dependency.
type catalogInfoYAML struct {
	Metadata struct {
		Name        string            `yaml:"name"`
		Description string            `yaml:"description"`
		Annotations map[string]string `yaml:"annotations"`
	} `yaml:"metadata"`
	Spec struct {
		Type      string `yaml:"type"`
		Lifecycle string `yaml:"lifecycle"`
		Owner     string `yaml:"owner"`
		System    string `yaml:"system"`
	} `yaml:"spec"`
}

type scorecardYAML struct {
	Service  string `yaml:"service"`
	Owner    string `yaml:"owner"`
	Criteria map[string]struct {
		Weight   int    `yaml:"weight"`
		Status   string `yaml:"status"`
		Evidence string `yaml:"evidence"`
	} `yaml:"criteria"`
	OverallScore int `yaml:"overall_score"`
}

type serviceRequestMetadata struct {
	Name string `json:"name"`
}

type valuesYAML struct {
	Image struct {
		Tag string `yaml:"tag"`
	} `yaml:"image"`
}

// LoadCatalog scans service-template-api/ and services/*/ for
// catalog-info.yaml (optionally paired with scorecard.yaml), the same
// discovery pavestack-portal/scripts/generate-catalog.mjs performs, and
// cross-references platform-config/tenants/<tenant>/{dev,prod}/values.yaml
// for the currently-pinned image tag per environment.
func LoadCatalog(repoRoot string) ([]Service, error) {
	var candidates []string

	if _, err := os.Stat(filepath.Join(repoRoot, "service-template-api", "catalog-info.yaml")); err == nil {
		candidates = append(candidates, filepath.Join(repoRoot, "service-template-api"))
	}

	servicesRoot := filepath.Join(repoRoot, "services")
	entries, err := os.ReadDir(servicesRoot)
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			dir := filepath.Join(servicesRoot, e.Name())
			if _, err := os.Stat(filepath.Join(dir, "catalog-info.yaml")); err == nil {
				candidates = append(candidates, dir)
			}
		}
	}

	services := make([]Service, 0, len(candidates))
	for _, dir := range candidates {
		svc, err := loadOne(repoRoot, dir)
		if err != nil {
			continue // a malformed catalog-info.yaml shouldn't 500 the whole catalog
		}
		services = append(services, svc)
	}

	sort.Slice(services, func(i, j int) bool { return services[i].Name < services[j].Name })
	return services, nil
}

// LoadOne resolves a single service by its catalog-info.yaml metadata.name.
func LoadOne(repoRoot, name string) (Service, bool) {
	all, err := LoadCatalog(repoRoot)
	if err != nil {
		return Service{}, false
	}
	for _, s := range all {
		if s.Name == name {
			return s, true
		}
	}
	return Service{}, false
}

func loadOne(repoRoot, dir string) (Service, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "catalog-info.yaml"))
	if err != nil {
		return Service{}, err
	}
	var ci catalogInfoYAML
	if err := yaml.Unmarshal(raw, &ci); err != nil {
		return Service{}, err
	}

	sc := scorecardYAML{OverallScore: -1}
	if raw, err := os.ReadFile(filepath.Join(dir, "scorecard.yaml")); err == nil {
		_ = yaml.Unmarshal(raw, &sc)
	}

	slug := ci.Metadata.Annotations["github.com/project-slug"]
	if slug == "" {
		slug = "pavestack/pavestack"
	}

	owner := ci.Spec.Owner
	if owner == "" {
		owner = sc.Owner
	}
	if owner == "" {
		owner = "unknown"
	}

	lifecycle := ci.Spec.Lifecycle
	if lifecycle == "" {
		lifecycle = "experimental"
	}

	createdVia := "manual"
	if ci.Metadata.Annotations["pavestack.io/created-via"] == "pave-cli" {
		createdVia = "pave-cli"
	}

	team := ci.Metadata.Annotations["pavestack.io/team"]
	if team == "" {
		team = owner
	}

	svc := Service{
		Name:         ci.Metadata.Name,
		Team:         team,
		Owner:        owner,
		Lifecycle:    lifecycle,
		System:       ci.Spec.System,
		Description:  ci.Metadata.Description,
		RepoURL:      "https://github.com/" + slug,
		CreatedVia:   createdVia,
		Tier:         ci.Metadata.Annotations["pavestack.io/tier"],
		Runtime:      ci.Metadata.Annotations["pavestack.io/runtime"],
		Exposure:     ci.Metadata.Annotations["pavestack.io/exposure"],
		Environments: map[string]EnvironmentStatus{},
	}

	if sc.OverallScore >= 0 {
		svc.Scorecard.OverallScore = sc.OverallScore
		keys := make([]string, 0, len(sc.Criteria))
		for k := range sc.Criteria {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			c := sc.Criteria[k]
			svc.Scorecard.Criteria = append(svc.Scorecard.Criteria, Criteria{
				Key:      k,
				Label:    strings.ReplaceAll(k, "_", " "),
				Weight:   c.Weight,
				Status:   c.Status,
				Evidence: c.Evidence,
			})
		}
	}

	tenant := resolveTenantName(dir, ci.Metadata.Name)
	for _, env := range []string{"dev", "prod"} {
		tag := readImageTag(filepath.Join(repoRoot, "platform-config", "tenants", tenant, env, "values.yaml"))
		if tag == "" {
			continue
		}
		// Sync/health are simulated pending a live Argo CD/Prometheus
		// integration - see decisions.md and the portal's own
		// generate-catalog.mjs, which makes the same disclosed
		// simplification today.
		svc.Environments[env] = EnvironmentStatus{Status: "synced", Health: "healthy", ImageTag: tag}
	}

	return svc, nil
}

func resolveTenantName(dir, catalogName string) string {
	if base := filepath.Base(dir); base == "service-template-api" {
		return "service-template-api"
	}

	meta, err := os.ReadFile(filepath.Join(dir, ".pavestack", "service-request.json"))
	if err == nil {
		var m serviceRequestMetadata
		if json.Unmarshal(meta, &m) == nil && m.Name != "" {
			return m.Name
		}
	}

	return strings.TrimSuffix(catalogName, "-api")
}

func readImageTag(valuesPath string) string {
	raw, err := os.ReadFile(valuesPath)
	if err != nil {
		return ""
	}
	var v valuesYAML
	if err := yaml.Unmarshal(raw, &v); err != nil {
		return ""
	}
	return v.Image.Tag
}
