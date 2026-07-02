locals {
  name = "${var.name_prefix}-${var.environment}"

  # Cost-attribution tag set - every AWS resource in this environment gets
  # these via merge(var.tags, ...) in each module (see modules/*/main.tf),
  # and the Team value matches the pavestack.io/team k8s label convention
  # so AWS Cost Explorer and in-cluster cost attribution use the same team
  # slug. See AGENTS.md "cost-tagging convention".
  tags = {
    Project     = "pavestack"
    Repository  = "platform-infra"
    Environment = var.environment
    ManagedBy   = "terraform"
    CostCenter  = var.cost_center
    Team        = var.team
  }

  github_actions_role_arns = var.enable_github_oidc_role ? [module.github_oidc[0].role_arn] : []
}

module "vpc" {
  source = "../../modules/vpc"

  name               = local.name
  vpc_cidr           = var.vpc_cidr
  az_count           = 2
  single_nat_gateway = true
  tags               = local.tags
}

module "ecr" {
  source = "../../modules/ecr"

  repositories = var.image_repositories
  tags         = local.tags
}

module "github_oidc" {
  count = var.enable_github_oidc_role ? 1 : 0

  source = "../../modules/github-oidc"

  name                 = local.name
  github_repository    = var.github_repository
  github_environment   = var.environment
  create_oidc_provider = var.create_github_oidc_provider
  tags                 = local.tags
}

module "eks" {
  source = "../../modules/eks"

  name                    = local.name
  kubernetes_version      = var.kubernetes_version
  vpc_id                  = module.vpc.vpc_id
  private_subnet_ids      = module.vpc.private_subnet_ids
  endpoint_public_access  = true
  endpoint_private_access = true
  node_instance_types     = ["t3.medium"]
  node_desired_size       = 2
  node_min_size           = 1
  node_max_size           = 3
  platform_admin_role_arns = setunion(
    var.platform_admin_role_arns,
    toset(local.github_actions_role_arns)
  )
  tags = local.tags
}

module "argocd" {
  source = "../../modules/argocd-bootstrap"

  chart_version = "9.5.17"

  depends_on = [module.eks]
}
