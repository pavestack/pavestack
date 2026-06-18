# Contributing to Pavestack

First off, thank you for considering contributing to Pavestack!

## How to Contribute

1. Fork the repository and create your feature branch from `main`.
2. Make your changes, following the architectural guidelines (all code must be production-aligned).
3. Ensure you have run `make test`, `make lint`, and `make fmt` locally.
4. Open a Pull Request detailing your changes.

## Development Workflow

- The project uses GitOps for delivery.
- Ensure all Terraform code is properly formatted via `make infra-fmt`.
- Security is a primary concern: Checkov, Trivy, and Gitleaks scans must pass in your PR.

We appreciate all community contributions!
