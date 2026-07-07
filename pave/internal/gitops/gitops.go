package gitops

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pavestack/pave/internal/cost"
	"github.com/pavestack/pave/internal/validate"
)

// WriteTenantManifests writes the initial ArgoCD tenant manifests by delegating to TenantManifestRenderer.
func WriteTenantManifests(repoRoot string, request validate.ServiceRequest, serviceDir string) error {
	if !validate.SafePathComponent(request.Name) {
		return fmt.Errorf("invalid service name %q", request.Name)
	}
	tenantRoot := filepath.Join(repoRoot, "platform-config", "tenants", request.Name)
	relHelmPath, err := filepath.Rel(repoRoot, filepath.Join(serviceDir, "deploy", "helm", request.Name+"-api"))
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(tenantRoot, "base"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(tenantRoot, "dev"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(tenantRoot, "prod"), 0o755); err != nil {
		return err
	}

	renderer := NewTenantManifestRenderer()
	profile := cost.ProfileFor(request.Tier)
	runtime := request.Runtime
	if runtime == "" {
		runtime = cost.DefaultRuntime
	}
	exposure := request.Exposure
	if exposure == "" {
		exposure = cost.DefaultExposure
	}

	tenantYAML, err := renderer.RenderTenantExtended(request.Name, filepath.ToSlash(relHelmPath), request.Team, request.Database, string(profile.Tier), runtime, exposure)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(tenantRoot, "tenant.yaml"), []byte(tenantYAML), 0o644); err != nil {
		return err
	}

	baseKustomization, err := renderer.RenderBaseKustomization(request.Name, request.Team)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(tenantRoot, "base", "kustomization.yaml"), []byte(baseKustomization), 0o644); err != nil {
		return err
	}

	imageRepo := fmt.Sprintf("123456789012.dkr.ecr.us-east-1.amazonaws.com/pavestack/%s-api", request.Name)
	resources := ResourceRequirements{
		RequestCPU:    profile.Resources.RequestCPU,
		RequestMemory: profile.Resources.RequestMemory,
		LimitCPU:      profile.Resources.LimitCPU,
		LimitMemory:   profile.Resources.LimitMemory,
	}
	devValues, err := renderer.RenderValuesTiered(request.Name, imageRepo, profile.Replicas.Dev, "debug", request.Team, resources)
	if err != nil {
		return err
	}
	prodValues, err := renderer.RenderValuesTiered(request.Name, imageRepo, profile.Replicas.Prod, "info", request.Team, resources)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(tenantRoot, "dev", "values.yaml"), []byte(devValues), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(tenantRoot, "prod", "values.yaml"), []byte(prodValues), 0o644); err != nil {
		return err
	}

	devApp, err := renderer.RenderApplication(request.Name, filepath.ToSlash(relHelmPath), "dev")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(tenantRoot, "dev", "application.yaml"), []byte(devApp), 0o644); err != nil {
		return err
	}

	prodApp, err := renderer.RenderApplication(request.Name, filepath.ToSlash(relHelmPath), "prod")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(tenantRoot, "prod", "application.yaml"), []byte(prodApp), 0o644); err != nil {
		return err
	}

	return nil
}

// CreatePullRequest delegates pull request creation to the VersionControl module.
func CreatePullRequest(repoRoot string, request validate.ServiceRequest, branch string) error {
	vc := NewVersionControl(repoRoot)
	return vc.CreatePullRequest(request, branch)
}

// CreatePullRequestURL delegates to VersionControl and also returns the
// created PR's URL, for callers (pave-api) that need to surface a link.
func CreatePullRequestURL(repoRoot string, request validate.ServiceRequest, branch string) (string, error) {
	vc := NewVersionControl(repoRoot)
	return vc.CreatePullRequestURL(request, branch)
}
