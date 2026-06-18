package gitops

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pavestack/pave/internal/validate"
)

func WriteTenantManifests(repoRoot string, request validate.ServiceRequest, serviceDir string) error {
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

	tenantYAML := fmt.Sprintf(`namespace: %s
helmPath: %s
releaseName: %s-api
owner: %s
database: %t
`, request.Name, filepath.ToSlash(relHelmPath), request.Name, request.Team, request.Database)
	if err := os.WriteFile(filepath.Join(tenantRoot, "tenant.yaml"), []byte(tenantYAML), 0o644); err != nil {
		return err
	}

	baseKustomization := fmt.Sprintf(`apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: %s
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
        value: %s
      - op: replace
        path: /metadata/labels/pavestack.io~1team
        value: %s
  - target:
      kind: Role
      name: developer
    patch: |
      - op: replace
        path: /metadata/namespace
        value: %s
  - target:
      kind: RoleBinding
      name: developer-binding
    patch: |
      - op: replace
        path: /metadata/namespace
        value: %s
      - op: replace
        path: /subjects/0/name
        value: %s
  - target:
      kind: NetworkPolicy
      name: default-deny
    patch: |
      - op: replace
        path: /metadata/namespace
        value: %s
  - target:
      kind: ResourceQuota
      name: tenant-quota
    patch: |
      - op: replace
        path: /metadata/namespace
        value: %s
`, request.Name, request.Name, request.Team, request.Name, request.Name, request.Team, request.Name, request.Name)

	if err := os.WriteFile(filepath.Join(tenantRoot, "base", "kustomization.yaml"), []byte(baseKustomization), 0o644); err != nil {
		return err
	}

	imageRepo := fmt.Sprintf("123456789012.dkr.ecr.us-east-1.amazonaws.com/pavestack/%s-api", request.Name)
	devValues := fmt.Sprintf(`replicaCount: 1

image:
  repository: %s
  tag: "0.1.0"

env:
  LOG_LEVEL: debug
  SERVICE_NAME: %s-api
`, imageRepo, request.Name)
	prodValues := fmt.Sprintf(`replicaCount: 2

image:
  repository: %s
  tag: "0.1.0"

env:
  LOG_LEVEL: info
  SERVICE_NAME: %s-api
`, imageRepo, request.Name)

	if err := os.WriteFile(filepath.Join(tenantRoot, "dev", "values.yaml"), []byte(devValues), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(tenantRoot, "prod", "values.yaml"), []byte(prodValues), 0o644); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(tenantRoot, "dev", "application.yaml"), []byte(renderApplication(request, relHelmPath, "dev")), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(tenantRoot, "prod", "application.yaml"), []byte(renderApplication(request, relHelmPath, "prod")), 0o644); err != nil {
		return err
	}

	return nil
}

func renderApplication(request validate.ServiceRequest, relHelmPath, environment string) string {
	return fmt.Sprintf(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: %s-api-%s
  namespace: argocd
  labels:
    pavestack.io/tenant: %s
    pavestack.io/environment: %s
spec:
  project: pavestack
  sources:
    - repoURL: https://github.com/pavestack/pavestack.git
      targetRevision: main
      ref: values
    - repoURL: https://github.com/pavestack/pavestack.git
      targetRevision: main
      path: %s
      helm:
        releaseName: %s-api
        valueFiles:
          - $values/platform-config/tenants/%s/%s/values.yaml
  destination:
    server: https://kubernetes.default.svc
    namespace: %s
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
`, request.Name, environment, request.Name, environment, filepath.ToSlash(relHelmPath), request.Name, request.Name, environment, request.Name)
}

func CreatePullRequest(repoRoot string, request validate.ServiceRequest, branch string) error {
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git not found in PATH")
	}
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("gh CLI not found; install GitHub CLI or pass --no-pr")
	}

	if branch == "" {
		branch = fmt.Sprintf("pave/create-%s-api", request.Name)
	}

	run := func(name string, args ...string) error {
		cmd := exec.Command(name, args...)
		cmd.Dir = repoRoot
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	if err := run("git", "checkout", "-b", branch); err != nil {
		return fmt.Errorf("create branch: %w", err)
	}

	paths := []string{
		filepath.Join("services", request.Name+"-api"),
		filepath.Join("platform-config", "tenants", request.Name),
	}

	if err := run("git", append([]string{"add"}, paths...)...); err != nil {
		return fmt.Errorf("git add: %w", err)
	}

	commitMsg := fmt.Sprintf("feat(%s): scaffold service via pave CLI", request.Name)
	if err := run("git", "commit", "-m", commitMsg); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	if err := run("git", "push", "-u", "origin", branch); err != nil {
		return fmt.Errorf("git push: %w", err)
	}

	title := fmt.Sprintf("feat(%s): scaffold %s-api", request.Name, request.Name)
	body := fmt.Sprintf(`Automated scaffold from pave create-service.

- Service: services/%s-api
- Tenant: platform-config/tenants/%s
- Owner: %s
- Database: %t

Argo CD reconciles after merge.`, request.Name, request.Name, request.Team, request.Database)

	return run("gh", "pr", "create", "--title", title, "--body", body)
}
