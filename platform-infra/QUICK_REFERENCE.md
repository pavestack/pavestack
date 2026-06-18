# Quick Reference Cheat Sheet

## One-Time Setup
```bash
# 1. Create state backend
cd bootstrap/remote-state
terraform init && terraform plan && terraform apply

# 2. Go to dev environment
cd ../../envs/dev

# 3. Set up config files
cp terraform.tfvars.example terraform.tfvars
cp backend.hcl.example backend.hcl
# ⚠️ Edit both files with your values
```

## Daily Commands (from repo root)
```bash
# Initialize Terraform
make init ENV=dev

# Check for errors
make validate ENV=dev

# Preview changes (dry run)
make plan ENV=dev

# Apply changes
make apply ENV=dev

# See what was created
terraform -chdir=envs/dev output

# Destroy everything ⚠️
make destroy ENV=dev

# Clean up temp files
make clean
```

## File Locations
| File | Purpose |
|------|---------|
| `envs/dev/terraform.tfvars` | Your configuration (copy from `.example`) |
| `envs/dev/backend.hcl` | Where to store state (copy from `.example`) |
| `envs/dev/main.tf` | Modules are called here |
| `modules/*/main.tf` | Actual infrastructure code |

## What Each Module Does (Quick)
- **network** → VPC, subnets, routing
- **eks** → Kubernetes cluster
- **ecr** → Docker image storage
- **github-oidc** → GitHub to AWS authentication (optional)
- **argocd-bootstrap** → Automatic deployment system

## Workflow
```
1. Configure terraform.tfvars and backend.hcl
   ↓
2. make init
   ↓
3. make plan (see what will change)
   ↓
4. make apply (create resources)
   ↓
5. Verify with: terraform -chdir=envs/dev output
```

## Don't Do This ❌
- Don't manually delete resources in AWS console
- Don't commit `terraform.tfvars` to Git
- Don't commit `.tfstate` files to Git
- Don't edit the generated `.tfplan` file
- Don't run terraform from different directories inconsistently

## Debug
```bash
# See what's in state
terraform -chdir=envs/dev state list

# Refresh state from AWS
terraform -chdir=envs/dev refresh

# See detailed info
terraform -chdir=envs/dev state show module.eks.aws_eks_cluster.main

# Check syntax
make validate ENV=dev

# Format code correctly
make fmt
```

## Useful Outputs
After `make apply`, run this to see important info:
```bash
terraform -chdir=envs/dev output
```

You'll see things like:
- Cluster name
- Cluster endpoint
- OIDC provider URL (for GitHub Actions)
- ECR repository URLs
