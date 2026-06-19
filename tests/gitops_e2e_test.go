package tests

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// parseYAMLSimple parses simple YAML with indentation.
// It returns a map of flat dot-separated keys, e.g. "image.repository": "..."
func parseYAMLSimple(content string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(content, "\n")
	var path []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Determine indentation depth
		indent := len(line) - len(strings.TrimLeft(line, " "))
		depth := indent / 2

		// Adjust path based on depth
		if depth < len(path) {
			path = path[:depth]
		}

		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		// If key starts with "- ", it's a list item.
		key = strings.TrimPrefix(key, "- ")

		currentPath := append(path, key)
		fullKey := strings.Join(currentPath, ".")

		if val != "" {
			result[fullKey] = val
		} else {
			// It's a parent key (nested block)
			if depth == len(path) {
				path = append(path, key)
			} else if depth < len(path) {
				path[depth] = key
			}
		}
	}
	return result
}

func getRepoRoot() (string, error) {
	if val := os.Getenv("PAVESTACK_ROOT"); val != "" {
		return val, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(cwd, "go.work")); err == nil {
			return cwd, nil
		}
		parent := filepath.Dir(cwd)
		if parent == cwd {
			break
		}
		cwd = parent
	}
	return "", fmt.Errorf("could not find repository root")
}

func TestGitOpsE2E(t *testing.T) {
	repoRoot, err := getRepoRoot()
	if err != nil {
		t.Fatalf("failed to find repository root: %v", err)
	}

	tempDir := t.TempDir()
	paveBin := filepath.Join(tempDir, "pave")

	t.Logf("Compiling pave CLI at %s...", paveBin)
	cmdBuild := exec.Command("go", "build", "-o", paveBin, "./pave/cmd/pave/main.go")
	cmdBuild.Dir = repoRoot
	if output, err := cmdBuild.CombinedOutput(); err != nil {
		t.Fatalf("failed to compile pave CLI: %v\nOutput:\n%s", err, output)
	}

	// Helper function to run pave command
	runPave := func(args ...string) (string, error) {
		cmd := exec.Command(paveBin, args...)
		cmd.Dir = repoRoot
		cmd.Env = append(os.Environ(), "PAVESTACK_ROOT="+repoRoot)
		output, err := cmd.CombinedOutput()
		return string(output), err
	}

	// Define test cases
	t.Run("CreateServiceSuccessfulValidationAndCleanup", func(t *testing.T) {
		svcName := "e2e-test-service"
		teamName := "e2e-test-team"

		// Cleanup function
		cleanup := func() {
			t.Log("Performing cleanup of generated files...")
			svcDir := filepath.Join(repoRoot, "services", svcName+"-api")
			tenantDir := filepath.Join(repoRoot, "platform-config", "tenants", svcName)
			os.RemoveAll(svcDir)
			os.RemoveAll(tenantDir)
		}
		// Clean before and after
		cleanup()
		t.Cleanup(cleanup)

		t.Log("Simulating service creation via pave create-service...")
		out, err := runPave("create-service", "--name", svcName, "--team", teamName, "--database=true", "--no-pr")
		if err != nil {
			t.Fatalf("pave create-service failed: %v\nOutput: %s", err, out)
		}

		t.Log("Verifying generated service directories and files...")
		svcDir := filepath.Join(repoRoot, "services", svcName+"-api")
		if _, err := os.Stat(svcDir); err != nil {
			t.Fatalf("expected service directory to exist at %s: %v", svcDir, err)
		}

		// Verify metadata file
		metadataPath := filepath.Join(svcDir, ".pavestack", "service-request.json")
		metadataBytes, err := os.ReadFile(metadataPath)
		if err != nil {
			t.Fatalf("failed to read service metadata: %v", err)
		}
		var meta map[string]interface{}
		if err := json.Unmarshal(metadataBytes, &meta); err != nil {
			t.Fatalf("failed to unmarshal service metadata: %v", err)
		}
		if meta["name"] != svcName || meta["team"] != teamName || meta["database"] != true {
			t.Errorf("unexpected service metadata: %+v", meta)
		}

		// Verify go.mod
		goModBytes, err := os.ReadFile(filepath.Join(svcDir, "go.mod"))
		if err != nil {
			t.Fatalf("failed to read go.mod: %v", err)
		}
		if !strings.Contains(string(goModBytes), "module github.com/pavestack/services/"+svcName+"-api") {
			t.Errorf("go.mod module name is incorrect: %s", string(goModBytes))
		}

		// Verify README database stub
		readmeBytes, err := os.ReadFile(filepath.Join(svcDir, "README.md"))
		if err != nil {
			t.Fatalf("failed to read README.md: %v", err)
		}
		if !strings.Contains(string(readmeBytes), "This service requested a managed database.") {
			t.Error("README.md should contain database stub")
		}

		// Verify Helm Chart.yaml
		chartBytes, err := os.ReadFile(filepath.Join(svcDir, "deploy", "helm", svcName+"-api", "Chart.yaml"))
		if err != nil {
			t.Fatalf("failed to read Helm Chart.yaml: %v", err)
		}
		if !strings.Contains(string(chartBytes), "name: "+svcName+"-api") {
			t.Errorf("Helm Chart.yaml has incorrect name: %s", string(chartBytes))
		}

		t.Log("Verifying generated GitOps tenant files...")
		tenantDir := filepath.Join(repoRoot, "platform-config", "tenants", svcName)
		if _, err := os.Stat(tenantDir); err != nil {
			t.Fatalf("expected tenant directory to exist at %s: %v", tenantDir, err)
		}

		// 1. Verify tenant.yaml
		tenantYAMLPath := filepath.Join(tenantDir, "tenant.yaml")
		tenantYAMLBytes, err := os.ReadFile(tenantYAMLPath)
		if err != nil {
			t.Fatalf("failed to read tenant.yaml: %v", err)
		}
		tenantFields := parseYAMLSimple(string(tenantYAMLBytes))
		if tenantFields["namespace"] != svcName {
			t.Errorf("expected namespace %s, got %s", svcName, tenantFields["namespace"])
		}
		if tenantFields["releaseName"] != svcName+"-api" {
			t.Errorf("expected releaseName %s-api, got %s", svcName, tenantFields["releaseName"])
		}
		if tenantFields["owner"] != teamName {
			t.Errorf("expected owner %s, got %s", teamName, tenantFields["owner"])
		}
		if tenantFields["database"] != "true" {
			t.Errorf("expected database true, got %s", tenantFields["database"])
		}

		// 2. Verify base/kustomization.yaml
		kustYAMLPath := filepath.Join(tenantDir, "base", "kustomization.yaml")
		kustYAMLBytes, err := os.ReadFile(kustYAMLPath)
		if err != nil {
			t.Fatalf("failed to read base/kustomization.yaml: %v", err)
		}
		kustContent := string(kustYAMLBytes)
		kustFields := parseYAMLSimple(kustContent)
		if kustFields["apiVersion"] != "kustomize.config.k8s.io/v1beta1" {
			t.Errorf("unexpected apiVersion in kustomization: %s", kustFields["apiVersion"])
		}
		if kustFields["kind"] != "Kustomization" {
			t.Errorf("unexpected kind in kustomization: %s", kustFields["kind"])
		}
		if kustFields["namespace"] != svcName {
			t.Errorf("unexpected namespace in kustomization: %s", kustFields["namespace"])
		}
		// String checks for resources and patches
		expectedKustSubstrings := []string{
			"../../../templates/namespace",
			"../../../templates/rbac",
			"../../../templates/network-policy",
			"../../../templates/resource-quota",
			"kind: Namespace",
			"name: REPLACE_NAMESPACE",
			"value: " + svcName,
			"value: " + teamName,
		}
		for _, sub := range expectedKustSubstrings {
			if !strings.Contains(kustContent, sub) {
				t.Errorf("base/kustomization.yaml missing expected substring: %q", sub)
			}
		}

		// 3. Verify dev/values.yaml and dev/application.yaml
		devValuesPath := filepath.Join(tenantDir, "dev", "values.yaml")
		devValuesBytes, err := os.ReadFile(devValuesPath)
		if err != nil {
			t.Fatalf("failed to read dev/values.yaml: %v", err)
		}
		devValuesFields := parseYAMLSimple(string(devValuesBytes))
		if devValuesFields["replicaCount"] != "1" {
			t.Errorf("expected replicaCount 1 in dev/values.yaml, got %s", devValuesFields["replicaCount"])
		}
		if devValuesFields["image.tag"] != "\"0.1.0\"" && devValuesFields["image.tag"] != "0.1.0" {
			t.Errorf("expected image tag 0.1.0 in dev/values.yaml, got %s", devValuesFields["image.tag"])
		}
		if devValuesFields["env.LOG_LEVEL"] != "debug" {
			t.Errorf("expected env.LOG_LEVEL debug in dev/values.yaml, got %s", devValuesFields["env.LOG_LEVEL"])
		}
		if devValuesFields["env.SERVICE_NAME"] != svcName+"-api" {
			t.Errorf("expected env.SERVICE_NAME %s-api in dev/values.yaml, got %s", svcName, devValuesFields["env.SERVICE_NAME"])
		}

		devAppPath := filepath.Join(tenantDir, "dev", "application.yaml")
		devAppBytes, err := os.ReadFile(devAppPath)
		if err != nil {
			t.Fatalf("failed to read dev/application.yaml: %v", err)
		}
		devAppContent := string(devAppBytes)
		devAppFields := parseYAMLSimple(devAppContent)
		if devAppFields["apiVersion"] != "argoproj.io/v1alpha1" {
			t.Errorf("expected dev application apiVersion argoproj.io/v1alpha1, got %s", devAppFields["apiVersion"])
		}
		if devAppFields["kind"] != "Application" {
			t.Errorf("expected dev application kind Application, got %s", devAppFields["kind"])
		}
		if devAppFields["metadata.name"] != svcName+"-api-dev" {
			t.Errorf("expected dev application name %s-api-dev, got %s", svcName, devAppFields["metadata.name"])
		}
		if devAppFields["metadata.labels.pavestack.io/environment"] != "dev" {
			t.Errorf("expected environment label dev, got %s", devAppFields["metadata.labels.pavestack.io/environment"])
		}
		if devAppFields["spec.syncPolicy.automated.prune"] != "true" {
			t.Errorf("expected automated prune true, got %s", devAppFields["spec.syncPolicy.automated.prune"])
		}
		if devAppFields["spec.syncPolicy.automated.selfHeal"] != "true" {
			t.Errorf("expected automated selfHeal true, got %s", devAppFields["spec.syncPolicy.automated.selfHeal"])
		}

		// 4. Verify prod/values.yaml and prod/application.yaml
		prodValuesPath := filepath.Join(tenantDir, "prod", "values.yaml")
		prodValuesBytes, err := os.ReadFile(prodValuesPath)
		if err != nil {
			t.Fatalf("failed to read prod/values.yaml: %v", err)
		}
		prodValuesFields := parseYAMLSimple(string(prodValuesBytes))
		if prodValuesFields["replicaCount"] != "2" {
			t.Errorf("expected replicaCount 2 in prod/values.yaml, got %s", prodValuesFields["replicaCount"])
		}
		if prodValuesFields["env.LOG_LEVEL"] != "info" {
			t.Errorf("expected env.LOG_LEVEL info in prod/values.yaml, got %s", prodValuesFields["env.LOG_LEVEL"])
		}

		prodAppPath := filepath.Join(tenantDir, "prod", "application.yaml")
		prodAppBytes, err := os.ReadFile(prodAppPath)
		if err != nil {
			t.Fatalf("failed to read prod/application.yaml: %v", err)
		}
		prodAppFields := parseYAMLSimple(string(prodAppBytes))
		if prodAppFields["metadata.name"] != svcName+"-api-prod" {
			t.Errorf("expected prod application name %s-api-prod, got %s", svcName, prodAppFields["metadata.name"])
		}
		if prodAppFields["metadata.labels.pavestack.io/environment"] != "prod" {
			t.Errorf("expected environment label prod, got %s", prodAppFields["metadata.labels.pavestack.io/environment"])
		}
	})

	t.Run("CLIValidationConstraints", func(t *testing.T) {
		invalidSvcNames := []string{
			"E2e-service",  // uppercase letter
			"1e2e-service", // starts with digit
			"-e2e-service", // starts with hyphen
			"e2e_service",  // underscore not allowed
			"e",            // too short (needs pattern ^[a-z][a-z0-9-]{1,48}[a-z0-9]$)
			"e2e-service-with-a-very-very-very-very-very-very-very-long-name-that-exceeds-fifty-characters", // too long
		}

		for _, name := range invalidSvcNames {
			t.Run("InvalidServiceName_"+name, func(t *testing.T) {
				out, err := runPave("create-service", "--name", name, "--team", "e2e-team", "--database=false", "--no-pr")
				if err == nil {
					t.Errorf("expected pave create-service to fail with invalid name %q but it succeeded. Output: %s", name, out)
				} else {
					t.Logf("Successfully failed with error: %v", err)
					if !strings.Contains(out, "validation failed") && !strings.Contains(out, "invalid") {
						t.Logf("Warning: Output did not contain expected validation error message: %s", out)
					}
				}
			})
		}

		invalidTeamNames := []string{
			"E2e-team", // uppercase
			"team_one", // underscore
			"team-with-a-very-very-very-very-very-very-very-very-very-very-very-very-very-very-long-name", // too long (>64 chars)
		}

		for _, team := range invalidTeamNames {
			t.Run("InvalidTeamName_"+team, func(t *testing.T) {
				out, err := runPave("create-service", "--name", "e2e-valid-svc", "--team", team, "--database=false", "--no-pr")
				if err == nil {
					t.Errorf("expected pave create-service to fail with invalid team %q but it succeeded. Output: %s", team, out)
				} else {
					t.Logf("Successfully failed with error: %v", err)
				}
			})
		}
	})

	t.Run("ServiceAlreadyExistsConstraint", func(t *testing.T) {
		svcName := "e2e-duplicate-svc"
		teamName := "e2e-team"

		cleanup := func() {
			svcDir := filepath.Join(repoRoot, "services", svcName+"-api")
			tenantDir := filepath.Join(repoRoot, "platform-config", "tenants", svcName)
			os.RemoveAll(svcDir)
			os.RemoveAll(tenantDir)
		}
		cleanup()
		t.Cleanup(cleanup)

		// First run: should succeed
		out, err := runPave("create-service", "--name", svcName, "--team", teamName, "--database=false", "--no-pr")
		if err != nil {
			t.Fatalf("first run failed unexpectedly: %v. Output: %s", err, out)
		}

		// Second run: should fail because it already exists
		out2, err2 := runPave("create-service", "--name", svcName, "--team", teamName, "--database=false", "--no-pr")
		if err2 == nil {
			t.Fatalf("expected second run to fail but it succeeded. Output: %s", out2)
		}
		t.Logf("Second run failed as expected: %v. Output: %s", err2, out2)

		if !strings.Contains(out2, "already exists") {
			t.Errorf("expected error output to contain 'already exists', got: %s", out2)
		}
	})
}
