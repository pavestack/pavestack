package gitops

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/pavestack/pave/internal/validate"
)

// gitCmdTimeout bounds each individual git/gh invocation so a hung process
// (e.g. gh waiting on an interactive prompt, or a stalled network call)
// can't wedge a create-service job - or, when driven by pave-api, an HTTP
// request goroutine - indefinitely.
const gitCmdTimeout = 30 * time.Second

// VersionControl encapsulates git and GitHub CLI operations.
type VersionControl struct {
	repoRoot string
}

// NewVersionControl creates a new VersionControl instance.
func NewVersionControl(repoRoot string) *VersionControl {
	return &VersionControl{
		repoRoot: repoRoot,
	}
}

// ValidateTools checks if the required tools (git and gh) are in the system's PATH.
func (vc *VersionControl) ValidateTools() error {
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git not found in PATH")
	}
	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("gh CLI not found; install GitHub CLI or pass --no-pr")
	}
	return nil
}

// CreatePullRequest checkout a branch, adds files, commits, pushes, and creates a pull request using the gh CLI.
func (vc *VersionControl) CreatePullRequest(request validate.ServiceRequest, branch string) error {
	_, err := vc.createPullRequest(request, branch, false)
	return err
}

// CreatePullRequestURL is identical to CreatePullRequest but also returns
// the created PR's URL (via `gh pr create --json url -q .url`), so a caller
// like the pave-api job runner can surface a clickable link.
func (vc *VersionControl) CreatePullRequestURL(request validate.ServiceRequest, branch string) (string, error) {
	return vc.createPullRequest(request, branch, true)
}

func (vc *VersionControl) createPullRequest(request validate.ServiceRequest, branch string, captureURL bool) (string, error) {
	if err := vc.ValidateTools(); err != nil {
		return "", err
	}

	if branch == "" {
		branch = fmt.Sprintf("pave/create-%s-api", request.Name)
	}

	run := func(name string, args ...string) error {
		ctx, cancel := context.WithTimeout(context.Background(), gitCmdTimeout)
		defer cancel()
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Dir = vc.repoRoot
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	if err := run("git", "checkout", "-b", branch); err != nil {
		return "", fmt.Errorf("create branch: %w", err)
	}

	paths := []string{
		filepath.Join("services", request.Name+"-api"),
		filepath.Join("platform-config", "tenants", request.Name),
	}

	if err := run("git", append([]string{"add"}, paths...)...); err != nil {
		return "", fmt.Errorf("git add: %w", err)
	}

	commitMsg := fmt.Sprintf("feat(%s): scaffold service via pave CLI", request.Name)
	if err := run("git", "commit", "-m", commitMsg); err != nil {
		return "", fmt.Errorf("git commit: %w", err)
	}

	if err := run("git", "push", "-u", "origin", branch); err != nil {
		return "", fmt.Errorf("git push: %w", err)
	}

	title := fmt.Sprintf("feat(%s): scaffold %s-api", request.Name, request.Name)
	body := fmt.Sprintf(`Automated scaffold from pave create-service.

- Service: services/%s-api
- Tenant: platform-config/tenants/%s
- Owner: %s
- Database: %t

Argo CD reconciles after merge.`, request.Name, request.Name, request.Team, request.Database)

	prArgs := []string{"pr", "create", "--title", title, "--body", body}
	if !captureURL {
		if err := run("gh", prArgs...); err != nil {
			return "", err
		}
		return "", nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), gitCmdTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "gh", append(prArgs, "--json", "url", "-q", ".url")...)
	cmd.Dir = vc.repoRoot
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
