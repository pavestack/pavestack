# Getting Started with Platform-Infra

This guide breaks down the overwhelming project into digestible pieces. Think of it as "what does each piece do, and why?"

## 🎯 The Big Picture

This project sets up a Kubernetes cluster in AWS with everything needed to deploy applications. It's split into **modules** (reusable building blocks) that are orchestrated in **environments** (dev, prod).

```
Your Infrastructure = Bootstrap + Network + EKS Cluster + Image Registry + Argo CD
```

---

## 📁 Folder Structure Explained

### 1. **`bootstrap/remote-state/`** — The Foundation
**What it does:** Creates the S3 bucket that securely stores your Terraform state (the "memory" of what infrastructure exists), utilizing S3-native state locking.

**When you use it:** Only once at the very beginning to set up state storage.

**Files:**
- `main.tf` — Creates S3 bucket (state storage) and KMS encryption key
- `variables.tf` — Configuration options (bucket name prefix, environment)

**Key Takeaway:** This is run separately and only once. Everything else depends on this existing.

---

### 2. **`modules/`** — The Building Blocks

Each module is a self-contained piece of infrastructure. Think of them as LEGO blocks.

#### **a) `modules/network`**
- **What:** VPC, subnets, Internet Gateway, NAT Gateway, routing
- **Why:** Provides the "plumbing" that EKS needs
- **Key inputs:** VPC CIDR (e.g., `10.0.0.0/16`), number of availability zones
- **Key outputs:** VPC ID, subnet IDs (used by other modules)

#### **b) `modules/eks`**
- **What:** The actual Kubernetes cluster, nodes, security groups, encryption, logging
- **Why:** This is where your applications will run
- **Key inputs:** VPC ID, subnet IDs (from network module), node size, node count
- **Key outputs:** Cluster name, API endpoint, OIDC provider (for GitHub Actions authentication)

#### **c) `modules/ecr`**
- **What:** Container image registries (Docker image storage)
- **Why:** To push and store Docker images for your applications
- **Key inputs:** List of repository names
- **Key outputs:** Repository URLs (used by applications to pull images)

#### **d) `modules/github-oidc`** (Optional)
- **What:** Authentication between GitHub Actions and AWS (no credentials needed!)
- **Why:** Secure way for GitHub Actions to deploy without storing AWS keys
- **Key inputs:** GitHub repository name, environment name
- **Key outputs:** IAM role ARN (that GitHub Actions can assume)

#### **e) `modules/argocd-bootstrap`**
- **What:** Installs Argo CD into the EKS cluster
- **Why:** Argo CD watches a Git repo and automatically deploys applications
- **Key inputs:** Helm chart version
- **Key outputs:** Argo CD namespace and connection info

---

### 3. **`envs/dev/` and `envs/prod/`** — Your Environments

These are where you **glue the modules together** for each environment.

**Dev environment (`envs/dev/`):**
- Smaller, cheaper cluster (for testing)
- Fewer nodes, smaller node sizes
- Used for development and testing

**Prod environment (`envs/prod/`):**
- Larger, redundant cluster (for real traffic)
- More nodes, larger node sizes
- High availability setup

**Key files in each environment:**
- `main.tf` — Calls all the modules and connects them together
- `variables.tf` — Configuration specific to that environment (VPC CIDR, node size, etc.)
- `terraform.tfvars.example` — Example values (copy this, fill it in, don't commit secrets!)
- `backend.hcl` — Points to the remote state bucket

---

### 4. **`scripts/terraform-env.sh`** — Helper Script

Sets up environment variables needed for Terraform to work correctly.

---

## 🚀 Step-by-Step Workflow

### Step 1: Setup State Backend (First Time Only)
```bash
cd bootstrap/remote-state
terraform init
terraform plan
terraform apply
```
**What happens:** Creates S3 bucket with native state locking and KMS encryption in AWS.
**You do this:** Once per AWS account, at the very beginning.

---

### Step 2: Configure Your Environment

Go to `envs/dev/` (or `envs/prod/`):

1. **Copy the example file:**
   ```bash
   cp terraform.tfvars.example terraform.tfvars
   ```

2. **Edit `terraform.tfvars`** and fill in:
   - `name_prefix` — Your project name
   - `vpc_cidr` — Network range (e.g., `10.0.0.0/16`)
   - `image_repositories` — Docker registries you need (e.g., `["platform-base", "tools"]`)
   - Other options based on environment needs

3. **Edit `backend.hcl`** and update the bucket name to match what you created in Step 1.

---

### Step 3: Initialize Terraform

```bash
# From the repo root:
make init ENV=dev
```

**What happens:** Downloads modules, connects to remote state, prepares to make changes.

---

### Step 4: Validate & Plan

```bash
# Check for syntax errors
make validate ENV=dev

# See what will be created (dry run)
make plan ENV=dev
```

**What happens:** Terraform shows you exactly what it will create/change before touching AWS.

---

### Step 5: Apply (Create Infrastructure)

```bash
make apply ENV=dev
```

**What happens:** Creates all the AWS resources. First time takes ~20 minutes.

---

### Step 6: Verify

```bash
# See the outputs (cluster name, endpoints, etc.)
terraform -chdir=envs/dev output
```

---

## 🎓 Key Concepts

### **Terraform State**
- Terraform tracks what it created in a `.tfstate` file (stored in S3)
- **Never** commit this to Git
- **Never** manually delete AWS resources (Terraform gets confused)
- If state gets out of sync, run `terraform refresh`

### **Variables vs. Outputs**
- **Variables:** Inputs to Terraform (what you configure)
- **Outputs:** Results from Terraform (what gets created, like cluster endpoints)
- Downstream apps read outputs to know how to connect

### **Modules**
- Modules are **reusable, self-contained** pieces
- Called from `main.tf` like: `module "network" { source = "../../modules/network" ... }`
- Each module has its own `variables.tf` (inputs) and `outputs.tf` (what it provides)

### **Environments**
- Dev and prod are separate so they don't interfere
- Both point to the same remote state backend
- Can have different sizes, configurations, security settings

---

## 🔧 Common Tasks

### Update a Module
1. Edit the module (e.g., `modules/eks/main.tf`)
2. Run `make plan ENV=dev` to see changes
3. Run `make apply ENV=dev` to apply them

### Scale the Cluster
1. Edit `envs/dev/terraform.tfvars`
2. Change `node_desired_size` or `node_instance_types`
3. Run `make plan ENV=dev` and `make apply ENV=dev`

### Add a New Image Repository
1. Edit `envs/dev/terraform.tfvars`
2. Add to `image_repositories` list
3. Run `make plan ENV=dev` and `make apply ENV=dev`

### Destroy Everything
```bash
make destroy ENV=dev
```
⚠️ Use with caution! This deletes all AWS resources.

---

## 🆘 Troubleshooting

| Problem | Solution |
|---------|----------|
| `Error: error reading S3 Bucket` | Check that `backend.hcl` bucket name is correct |
| `Error: failed to fetch state` | Run `make init ENV=dev` again |
| State is out of sync | Run `terraform -chdir=envs/dev refresh` |
| Module changes don't apply | Run `make clean` and `make init ENV=dev` |

---

## 📚 Next Steps

1. **Understand your network:** Review `modules/network/outputs.tf` to see what subnets exist
2. **Connect to the cluster:** Once EKS is created, AWS documentation shows how to configure `kubectl`
3. **Check Argo CD:** The `argocd-bootstrap` module installs it; check `modules/argocd-bootstrap/outputs.tf`
4. **Deploy applications:** Create a separate GitOps repository that Argo CD watches

---

## 💡 Mental Model

Think of this project as a **restaurant kitchen setup:**

- **Bootstrap** = Installing plumbing and electrical (state backend)
- **Network module** = Setting up the building structure (VPC, walls, rooms)
- **EKS module** = Installing the kitchen equipment (stove, ovens, prep surfaces)
- **ECR module** = Setting up ingredient storage (refrigerators, pantry)
- **Argo CD** = Hiring a manager to oversee operations (automatically handles deployments)
- **Environment files** = Different kitchen sizes for a food truck (dev) vs. a restaurant (prod)

Once the kitchen is ready, application teams (other repos) cook meals (deploy apps) using Argo CD!
