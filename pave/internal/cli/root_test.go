package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRepoRootEnvVar(t *testing.T) {
	t.Setenv("PAVESTACK_ROOT", "/my/custom/root")
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if root != "/my/custom/root" {
		t.Errorf("expected root '/my/custom/root', got %q", root)
	}
}

func TestRepoRootCwdTraverse(t *testing.T) {
	// Unset PAVESTACK_ROOT to force traversing
	t.Setenv("PAVESTACK_ROOT", "")

	temp := t.TempDir()
	// Create mock pavestack workspace
	if err := os.MkdirAll(filepath.Join(temp, "platform-config"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(temp, "service-template-api"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(temp, "pave"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a deep subdirectory to traverse from
	deepDir := filepath.Join(temp, "pave", "internal", "cli")
	if err := os.MkdirAll(deepDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatal(err)
		}
	})

	if err := os.Chdir(deepDir); err != nil {
		t.Fatal(err)
	}

	root, err := repoRoot()
	if err != nil {
		t.Fatalf("expected to find repo root by traversing, got %v", err)
	}

	// Clean paths to compare
	expected, err := filepath.EvalSymlinks(temp)
	if err != nil {
		expected = temp
	}
	actual, err := filepath.EvalSymlinks(root)
	if err != nil {
		actual = root
	}

	if actual != expected {
		t.Errorf("expected repo root %q, got %q", expected, actual)
	}
}

func TestRepoRootNotFound(t *testing.T) {
	t.Setenv("PAVESTACK_ROOT", "")

	temp := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(cwd); err != nil {
			t.Fatal(err)
		}
	})

	if err := os.Chdir(temp); err != nil {
		t.Fatal(err)
	}

	_, err = repoRoot()
	if err == nil {
		t.Fatal("expected error finding repo root in empty directory")
	}
}
