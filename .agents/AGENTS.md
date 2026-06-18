# Pavestack Agent Guidelines

When working in this repository, you must adhere to the following project-specific rules, architectural patterns, and security constraints.

## 1. Monorepo Architecture
This is an Internal Developer Platform (IDP) monorepo. Understand the directory boundaries:
*   `platform-infra/`: Core Terraform modules (EKS, VPC, OIDC, ArgoCD).
*   `platform-config/`: GitOps manifests (reconciled by ArgoCD).
*   `service-template-api/`: The golden-path Go API scaffold.
*   `pave/`: The self-service CLI.
*   `pavestack-portal/`: The developer portal (Next.js/React).
*   `services/`: Instantiated microservices.

## 2. GitOps Delivery Model
*   **No Direct Applies**: Never run `kubectl apply` to deploy applications. All deployments and state changes must be committed as declarative manifests to the `platform-config/` directory so that Argo CD can reconcile them.
*   **Infrastructure as Code**: All infrastructure changes must be made via Terraform in `platform-infra/`. 

## 3. Security by Default
*   **Containers**: Use Distroless base images and ensure containers run as `nonroot`. Ensure Dockerfiles include a `HEALTHCHECK` instruction (or `HEALTHCHECK NONE`).
*   **Networking**: Ensure default-deny NetworkPolicies are respected.
*   **Secrets**: Never commit secrets. Rely on GitHub OIDC for CI operations.

## 4. CI/CD & Testing
*   The `main` branch is protected. All changes must pass strict CI checks.
*   **Security Scanners**: Your code must satisfy Checkov, Trivy, and TFLint rules. If you encounter Checkov errors in GitHub Actions workflows (like `CKV_GHA_7`), correctly document the skips or fix the underlying issue.
*   **Local Checks**: Always use the provided `Makefile` targets to validate changes locally before committing:
    *   Go code: `make fmt`, `make lint`, `make test`
    *   Terraform code: `make infra-fmt`, `make infra-validate`
*   **Commit Messages**: We use Release Please for automated versioning. You **MUST** use Conventional Commits (e.g., `feat:`, `fix:`, `chore:`, `docs:`) for all commit messages so the changelog generates correctly.

## 5. Dependabot Findings
*   When fixing dependabot findings that update GitHub Actions workflows, `gh pr merge` will fail due to the `gh` CLI lacking the `workflow` OAuth scope.
*   Instead of merging directly, use the `gh` CLI to instruct Dependabot to merge the PR: `gh pr comment <pr-number> -b "@dependabot merge"`. Dependabot has the proper permissions to update workflows.
