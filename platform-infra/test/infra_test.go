package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// PlanOutput represents the structured JSON output from terraform show -json.
type PlanOutput struct {
	PlannedValues PlannedValues `json:"planned_values"`
}

type PlannedValues struct {
	RootModule RootModule `json:"root_module"`
}

type RootModule struct {
	Resources    []Resource    `json:"resources"`
	ChildModules []ChildModule `json:"child_modules"`
}

type ChildModule struct {
	Address      string        `json:"address"`
	Resources    []Resource    `json:"resources"`
	ChildModules []ChildModule `json:"child_modules"`
}

type Resource struct {
	Address string                 `json:"address"`
	Type    string                 `json:"type"`
	Name    string                 `json:"name"`
	Values  map[string]interface{} `json:"values"`
}

// findResource searches child modules recursively for a resource with the given address.
func findResource(module ChildModule, address string) (Resource, bool) {
	for _, res := range module.Resources {
		if res.Address == address {
			return res, true
		}
	}
	for _, child := range module.ChildModules {
		if res, found := findResource(child, address); found {
			return res, true
		}
	}
	return Resource{}, false
}

// findResourceInRoot searches the entire plan for a resource with the given address.
func findResourceInRoot(plan PlanOutput, address string) (Resource, bool) {
	for _, res := range plan.PlannedValues.RootModule.Resources {
		if res.Address == address {
			return res, true
		}
	}
	for _, child := range plan.PlannedValues.RootModule.ChildModules {
		if res, found := findResource(child, address); found {
			return res, true
		}
	}
	return Resource{}, false
}

// copyDir copies src directory contents to dst, excluding git and state files.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if info.IsDir() {
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			return os.MkdirAll(target, info.Mode())
		}

		if info.Name() == "terraform.tfstate" || info.Name() == "terraform.tfstate.backup" || strings.HasSuffix(info.Name(), ".tfstate") {
			return nil
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		return err
	})
}

// startMockAWSServer sets up a local HTTP server mock for EC2, STS, and EKS.
func startMockAWSServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(r.Body)
		body := string(bodyBytes)

		if strings.Contains(body, "Action=GetCallerIdentity") {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?>
<GetCallerIdentityResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
  <GetCallerIdentityResult>
    <Arn>arn:aws:iam::123456789012:user/mock</Arn>
    <UserId>AKIAIOSFODNN7EXAMPLE</UserId>
    <Account>123456789012</Account>
  </GetCallerIdentityResult>
  <ResponseMetadata>
    <RequestId>01234567-89ab-cdef-0123-456789abcdef</RequestId>
  </ResponseMetadata>
</GetCallerIdentityResponse>`))
			return
		}

		if strings.Contains(body, "Action=DescribeAvailabilityZones") {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?>
<DescribeAvailabilityZonesResponse xmlns="http://ec2.amazonaws.com/doc/2016-11-15/">
  <requestId>01234567-89ab-cdef-0123-456789abcdef</requestId>
  <availabilityZoneInfo>
    <item>
      <zoneName>eu-central-1a</zoneName>
      <zoneState>available</zoneState>
      <regionName>eu-central-1</regionName>
      <optInStatus>opt-in-not-required</optInStatus>
    </item>
    <item>
      <zoneName>eu-central-1b</zoneName>
      <zoneState>available</zoneState>
      <regionName>eu-central-1</regionName>
      <optInStatus>opt-in-not-required</optInStatus>
    </item>
    <item>
      <zoneName>eu-central-1c</zoneName>
      <zoneState>available</zoneState>
      <regionName>eu-central-1</regionName>
      <optInStatus>opt-in-not-required</optInStatus>
    </item>
  </availabilityZoneInfo>
</DescribeAvailabilityZonesResponse>`))
			return
		}

		if strings.Contains(r.URL.Path, "/clusters") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
  "cluster": {
    "name": "pavestack-mock",
    "arn": "arn:aws:eks:eu-central-1:123456789012:cluster/pavestack-mock",
    "endpoint": "https://localhost:12345",
    "certificateAuthority": {
      "data": "dGVzdC1jZXJ0aWZpY2F0ZS1hdXRob3JpdHktZGF0YQ=="
    },
    "identity": {
      "oidc": {
        "issuer": "https://localhost:12345/oidc"
      }
    },
    "status": "ACTIVE"
  }
}`))
			return
		}

		// Default fallback
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?><Response><Message>Mock OK</Message></Response>`))
	}))
}

// runTerraform runs a terraform command in the specified directory.
func runTerraform(t *testing.T, dir string, args ...string) (string, error) {
	cmd := exec.Command("terraform", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"AWS_ACCESS_KEY_ID=mock_key",
		"AWS_SECRET_ACCESS_KEY=mock_secret",
		"AWS_DEFAULT_REGION=eu-central-1",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("terraform %s failed: %w\nOutput:\n%s", strings.Join(args, " "), err, out)
	}
	return string(out), nil
}

func getIntValue(v interface{}) (int, bool) {
	switch val := v.(type) {
	case float64:
		return int(val), true
	case int:
		return val, true
	}
	return 0, false
}

func testEnvironment(t *testing.T, envName, expectedCidr, expectedInstanceType string, expectedDesiredSize, expectedFlowLogRetention int) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working dir: %v", err)
	}

	// The working directory for the test package is platform-infra/test
	// The parent of that is platform-infra/
	platformInfraDir := filepath.Dir(wd)

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "platform-infra-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Copy platform-infra to tempDir
	err = copyDir(platformInfraDir, tempDir)
	if err != nil {
		t.Fatalf("failed to copy platform-infra: %v", err)
	}

	// Delete backend.tf in the target env folder to force local backend
	targetEnvDir := filepath.Join(tempDir, "envs", envName)
	backendFile := filepath.Join(targetEnvDir, "backend.tf")
	if err := os.Remove(backendFile); err != nil && !os.IsNotExist(err) {
		t.Fatalf("failed to remove backend.tf: %v", err)
	}

	// Start AWS Mock Server
	server := startMockAWSServer(t)
	defer server.Close()

	// Write providers_override.tf configuring AWS provider to local mock URL
	overrideContent := fmt.Sprintf(`
provider "aws" {
  region                      = "eu-central-1"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  access_key                  = "mock_key"
  secret_key                  = "mock_secret"

  endpoints {
    ec2 = "%s"
    sts = "%s"
    eks = "%s"
  }
}
`, server.URL, server.URL, server.URL)

	overrideFile := filepath.Join(targetEnvDir, "providers_override.tf")
	if err := os.WriteFile(overrideFile, []byte(overrideContent), 0644); err != nil {
		t.Fatalf("failed to write providers_override.tf: %v", err)
	}

	// Initialize Terraform
	t.Logf("Running terraform init in %s...", targetEnvDir)
	initOut, err := runTerraform(t, targetEnvDir, "init", "-backend=false")
	if err != nil {
		t.Fatalf("init failed: %v\nOutput: %s", err, initOut)
	}

	// Run Terraform Plan
	t.Logf("Running terraform plan in %s...", targetEnvDir)
	planOut, err := runTerraform(t, targetEnvDir, "plan", "-out=tfplan", "-input=false")
	if err != nil {
		t.Fatalf("plan failed: %v\nOutput: %s", err, planOut)
	}

	// Generate JSON plan representation
	showOut, err := runTerraform(t, targetEnvDir, "show", "-json", "tfplan")
	if err != nil {
		t.Fatalf("terraform show failed: %v\nOutput: %s", err, showOut)
	}

	// Parse plan JSON
	var plan PlanOutput
	if err := json.Unmarshal([]byte(showOut), &plan); err != nil {
		t.Fatalf("failed to parse plan JSON: %v", err)
	}

	// 1. Verify VPC resource configuration
	vpcRes, found := findResourceInRoot(plan, "module.vpc.aws_vpc.this")
	if !found {
		t.Fatalf("module.vpc.aws_vpc.this resource not found in plan")
	}

	cidr, ok := vpcRes.Values["cidr_block"].(string)
	if !ok {
		t.Fatalf("cidr_block not found or not a string in aws_vpc.this resource values")
	}
	if cidr != expectedCidr {
		t.Errorf("expected VPC CIDR %s, got %s", expectedCidr, cidr)
	}

	// 2. Verify EKS Node Group configuration
	nodeGroupRes, found := findResourceInRoot(plan, "module.eks.aws_eks_node_group.default")
	if !found {
		t.Fatalf("module.eks.aws_eks_node_group.default resource not found in plan")
	}

	// Check instance types
	instanceTypes, ok := nodeGroupRes.Values["instance_types"].([]interface{})
	if !ok || len(instanceTypes) == 0 {
		t.Fatalf("instance_types not found or empty in aws_eks_node_group.default")
	}
	if instanceTypes[0] != expectedInstanceType {
		t.Errorf("expected node instance type %s, got %v", expectedInstanceType, instanceTypes[0])
	}

	// Check scaling config desired size
	scalingConfig, ok := nodeGroupRes.Values["scaling_config"].([]interface{})
	if !ok || len(scalingConfig) == 0 {
		t.Fatalf("scaling_config not found or empty in aws_eks_node_group.default")
	}
	scMap, ok := scalingConfig[0].(map[string]interface{})
	if !ok {
		t.Fatalf("scaling_config[0] is not a map")
	}
	desiredSize, ok := getIntValue(scMap["desired_size"])
	if !ok {
		t.Fatalf("desired_size not found or not an integer in scaling_config")
	}
	if desiredSize != expectedDesiredSize {
		t.Errorf("expected scaling desired size %d, got %d", expectedDesiredSize, desiredSize)
	}

	// 3. Verify VPC flow logs and their CloudWatch log group, with per-env retention.
	flowLogGroup, found := findResourceInRoot(plan, "module.vpc.aws_cloudwatch_log_group.flow_logs[0]")
	if !found {
		t.Fatalf("module.vpc.aws_cloudwatch_log_group.flow_logs[0] resource not found in plan")
	}
	retention, ok := getIntValue(flowLogGroup.Values["retention_in_days"])
	if !ok {
		t.Fatalf("retention_in_days not found or not an integer in flow-log CloudWatch log group")
	}
	if retention != expectedFlowLogRetention {
		t.Errorf("expected flow-log retention_in_days %d, got %d", expectedFlowLogRetention, retention)
	}

	// Verify modules presence
	expectedResources := []string{
		"module.vpc.aws_vpc.this",
		"module.eks.aws_eks_cluster.this",
		"module.eks.aws_eks_node_group.default",
		// VPC flow logs
		"module.vpc.aws_flow_log.this[0]",
		"module.vpc.aws_cloudwatch_log_group.flow_logs[0]",
		// observability stack
		"module.observability.helm_release.kube_prometheus_stack",
		"module.observability.helm_release.loki",
		"module.observability.helm_release.tempo",
		// ingress: AWS Load Balancer Controller + its IRSA role (external-dns is
		// gated on route53_zone_id, which is empty in plan-only test runs)
		"module.ingress.helm_release.aws_load_balancer_controller[0]",
		"module.ingress.aws_iam_role.aws_load_balancer_controller[0]",
	}
	for _, resAddr := range expectedResources {
		if _, found := findResourceInRoot(plan, resAddr); !found {
			t.Errorf("expected resource %s to be planned but it was not found", resAddr)
		}
	}

	t.Logf("All validations passed for %s environment!", envName)
}

func TestDevEnvironment(t *testing.T) {
	testEnvironment(t, "dev", "10.20.0.0/16", "t3.medium", 2, 14)
}

func TestProdEnvironment(t *testing.T) {
	testEnvironment(t, "prod", "10.30.0.0/16", "m6i.large", 3, 90)
}
