locals {
  name = "${var.name_prefix}-${var.environment}"

  tags = {
    Project     = "pavestack"
    Repository  = "platform-infra"
    Environment = var.environment
    ManagedBy   = "terraform"
  }

  github_actions_role_arns = var.enable_github_oidc_role ? [module.github_oidc[0].role_arn] : []
}

module "vpc" {
  source = "../../modules/vpc"

  name               = local.name
  vpc_cidr           = var.vpc_cidr
  az_count           = 3
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
  node_instance_types     = ["m6i.large"]
  node_desired_size       = 3
  node_min_size           = 3
  node_max_size           = 6
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
