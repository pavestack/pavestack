package gitops

import (
	"bytes"
	"text/template"
)

// Define templates as constants
const (
	tenantTemplate = `namespace: {{.Name}}
helmPath: {{.RelHelmPath}}
releaseName: {{.Name}}-api
owner: {{.Team}}
database: {{.Database}}
`

	baseKustomizationTemplate = `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: {{.Name}}
resources:
  - ../../../templates/namespace
  - ../../../templates/rbac
  - ../../../templates/network-policy
  - ../../../templates/resource-quota
patches:
  - target:
      kind: Namespace
      name: REPLACE_NAMESPACE
    patch: |
      - op: replace
        path: /metadata/name
        value: {{.Name}}
      - op: replace
        path: /metadata/labels/pavestack.io~1team
        value: {{.Team}}
  - target:
      kind: Role
      name: developer
    patch: |
      - op: replace
        path: /metadata/namespace
        value: {{.Name}}
  - target:
      kind: RoleBinding
      name: developer-binding
    patch: |
      - op: replace
        path: /metadata/namespace
        value: {{.Name}}
      - op: replace
        path: /subjects/0/name
        value: {{.Team}}
  - target:
      kind: NetworkPolicy
      name: default-deny
    patch: |
      - op: replace
        path: /metadata/namespace
        value: {{.Name}}
  - target:
      kind: NetworkPolicy
      name: allow-egress-dns
    patch: |
      - op: replace
        path: /metadata/namespace
        value: {{.Name}}
  - target:
      kind: ResourceQuota
      name: tenant-quota
    patch: |
      - op: replace
        path: /metadata/namespace
        value: {{.Name}}
  - target:
      kind: LimitRange
      name: tenant-limits
    patch: |
      - op: replace
        path: /metadata/namespace
        value: {{.Name}}
`

	valuesTemplate = `replicaCount: {{.ReplicaCount}}

image:
  repository: {{.ImageRepo}}
  tag: "0.1.0"

env:
  LOG_LEVEL: {{.LogLevel}}
  SERVICE_NAME: {{.Name}}-api
`

	// tenantTemplateExtended supersedes tenantTemplate for tenants created
	// after tier/runtime/exposure were introduced. tenantTemplate is kept
	// so RenderTenant's existing contract (and its tests) stay stable.
	tenantTemplateExtended = `namespace: {{.Name}}
helmPath: {{.RelHelmPath}}
releaseName: {{.Name}}-api
owner: {{.Team}}
database: {{.Database}}
tier: {{.Tier}}
runtime: {{.Runtime}}
exposure: {{.Exposure}}
`

	// valuesTemplateTiered supersedes valuesTemplate to also pin the
	// per-environment resource requests/limits derived from the service's
	// tier, so tier sizing is enforced by the values Argo CD actually
	// applies rather than left to chart defaults alone.
	valuesTemplateTiered = `replicaCount: {{.ReplicaCount}}

image:
  repository: {{.ImageRepo}}
  tag: "0.1.0"

# team feeds the pavestack.io/team label the require-pavestack-labels
# Kyverno ClusterPolicy enforces on every workload (see
# platform-config/policies/kyverno/require-labels.yaml).
team: {{.Team}}

resources:
  requests:
    cpu: {{.RequestCPU}}
    memory: {{.RequestMemory}}
  limits:
    cpu: {{.LimitCPU}}
    memory: {{.LimitMemory}}

env:
  LOG_LEVEL: {{.LogLevel}}
  SERVICE_NAME: {{.Name}}-api
`

	applicationTemplate = `apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: {{.Name}}-api-{{.Environment}}
  namespace: argocd
  labels:
    pavestack.io/tenant: {{.Name}}
    pavestack.io/environment: {{.Environment}}
spec:
  project: pavestack
  sources:
    - repoURL: https://github.com/pavestack/pavestack.git
      targetRevision: main
      ref: values
    - repoURL: https://github.com/pavestack/pavestack.git
      targetRevision: main
      path: {{.RelHelmPath}}
      helm:
        releaseName: {{.Name}}-api
        valueFiles:
          - $values/platform-config/tenants/{{.Name}}/{{.Environment}}/values.yaml
  destination:
    server: https://kubernetes.default.svc
    namespace: {{.Name}}
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
`
)

// TenantManifestRenderer encapsulates the rendering logic for tenant manifests using Go text/template.
type TenantManifestRenderer struct{}

// NewTenantManifestRenderer creates a new TenantManifestRenderer.
func NewTenantManifestRenderer() *TenantManifestRenderer {
	return &TenantManifestRenderer{}
}

// We parse templates once at package load time.
var (
	tTenant            = template.Must(template.New("tenant").Parse(tenantTemplate))
	tTenantExtended    = template.Must(template.New("tenantExtended").Parse(tenantTemplateExtended))
	tBaseKustomization = template.Must(template.New("baseKustomization").Parse(baseKustomizationTemplate))
	tValues            = template.Must(template.New("values").Parse(valuesTemplate))
	tValuesTiered      = template.Must(template.New("valuesTiered").Parse(valuesTemplateTiered))
	tApplication       = template.Must(template.New("application").Parse(applicationTemplate))
)

// Template data structures
type tenantData struct {
	Name        string
	RelHelmPath string
	Team        string
	Database    bool
}

type tenantDataExtended struct {
	Name        string
	RelHelmPath string
	Team        string
	Database    bool
	Tier        string
	Runtime     string
	Exposure    string
}

type baseKustomizationData struct {
	Name string
	Team string
}

type valuesData struct {
	Name         string
	ImageRepo    string
	ReplicaCount int
	LogLevel     string
}

type valuesDataTiered struct {
	Name          string
	ImageRepo     string
	ReplicaCount  int
	LogLevel      string
	Team          string
	RequestCPU    string
	RequestMemory string
	LimitCPU      string
	LimitMemory   string
}

type applicationData struct {
	Name        string
	Environment string
	RelHelmPath string
}

// RenderTenant returns the rendered tenant.yaml content.
func (r *TenantManifestRenderer) RenderTenant(name, relHelmPath, team string, database bool) (string, error) {
	data := tenantData{
		Name:        name,
		RelHelmPath: relHelmPath,
		Team:        team,
		Database:    database,
	}
	var buf bytes.Buffer
	if err := tTenant.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderTenantExtended returns the rendered tenant.yaml content including
// tier/runtime/exposure. New tenants are always written with this method;
// RenderTenant is kept only for its existing tested contract.
func (r *TenantManifestRenderer) RenderTenantExtended(name, relHelmPath, team string, database bool, tier, runtime, exposure string) (string, error) {
	data := tenantDataExtended{
		Name:        name,
		RelHelmPath: relHelmPath,
		Team:        team,
		Database:    database,
		Tier:        tier,
		Runtime:     runtime,
		Exposure:    exposure,
	}
	var buf bytes.Buffer
	if err := tTenantExtended.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderBaseKustomization returns the rendered base/kustomization.yaml content.
func (r *TenantManifestRenderer) RenderBaseKustomization(name, team string) (string, error) {
	data := baseKustomizationData{
		Name: name,
		Team: team,
	}
	var buf bytes.Buffer
	if err := tBaseKustomization.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderValues returns the rendered values.yaml content.
func (r *TenantManifestRenderer) RenderValues(name, imageRepo string, replicaCount int, logLevel string) (string, error) {
	data := valuesData{
		Name:         name,
		ImageRepo:    imageRepo,
		ReplicaCount: replicaCount,
		LogLevel:     logLevel,
	}
	var buf bytes.Buffer
	if err := tValues.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderValuesTiered returns rendered values.yaml content including the
// resource requests/limits for the service's tier. New tenants are always
// written with this method; RenderValues is kept only for its existing
// tested contract.
func (r *TenantManifestRenderer) RenderValuesTiered(name, imageRepo string, replicaCount int, logLevel, team string, resources ResourceRequirements) (string, error) {
	data := valuesDataTiered{
		Name:          name,
		ImageRepo:     imageRepo,
		ReplicaCount:  replicaCount,
		LogLevel:      logLevel,
		Team:          team,
		RequestCPU:    resources.RequestCPU,
		RequestMemory: resources.RequestMemory,
		LimitCPU:      resources.LimitCPU,
		LimitMemory:   resources.LimitMemory,
	}
	var buf bytes.Buffer
	if err := tValuesTiered.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ResourceRequirements mirrors cost.ResourceProfile without importing the
// cost package here, keeping tenant_manifest.go a pure rendering module.
type ResourceRequirements struct {
	RequestCPU    string
	RequestMemory string
	LimitCPU      string
	LimitMemory   string
}

// RenderApplication returns the rendered application.yaml content.
func (r *TenantManifestRenderer) RenderApplication(name, relHelmPath, environment string) (string, error) {
	data := applicationData{
		Name:        name,
		Environment: environment,
		RelHelmPath: relHelmPath,
	}
	var buf bytes.Buffer
	if err := tApplication.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
