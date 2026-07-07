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

  name                    = local.name
  vpc_cidr                = var.vpc_cidr
  az_count                = 3
  single_nat_gateway      = true
  flow_log_retention_days = 30
  tags                    = local.tags
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
  # Minimal static "system" pool sized like dev; staging mirrors prod's
  # topology but runs dev-grade capacity.
  node_instance_types = ["t3.medium"]
  node_desired_size   = 2
  node_min_size       = 1
  node_max_size       = 3
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

module "observability" {
  source = "../../modules/observability"

  prometheus_storage_size = "10Gi"
  prometheus_retention    = "7d"
  alert_webhook_url       = var.alert_webhook_url

  depends_on = [module.eks]
}

module "ingress" {
  source = "../../modules/ingress"

  cluster_name      = module.eks.cluster_name
  region            = var.aws_region
  vpc_id            = module.vpc.vpc_id
  oidc_provider_arn = module.eks.oidc_provider_arn
  oidc_issuer_url   = module.eks.cluster_oidc_issuer_url
  route53_zone_id   = var.route53_zone_id
  domain_filter     = var.platform_domain
  # external-dns needs a real hosted zone; its IAM policy scopes to route53_zone_id
  # and an empty zone id would produce a broken (but plannable) role. Gate it on a
  # zone id being provided. Real deployments set route53_zone_id via tfvars.
  enable_external_dns = var.route53_zone_id != ""
  tags                = local.tags

  depends_on = [module.eks]
}

module "secrets" {
  source = "../../modules/secrets"

  cluster_name      = module.eks.cluster_name
  region            = var.aws_region
  oidc_provider_arn = module.eks.oidc_provider_arn
  oidc_issuer_url   = module.eks.cluster_oidc_issuer_url
  tags              = local.tags

  depends_on = [module.eks]
}

module "policy" {
  source = "../../modules/policy"

  chart_version = "3.2.6"

  depends_on = [module.eks]
}

module "karpenter" {
  source = "../../modules/karpenter"

  cluster_name      = module.eks.cluster_name
  cluster_endpoint  = module.eks.cluster_endpoint
  region            = var.aws_region
  oidc_provider_arn = module.eks.oidc_provider_arn
  oidc_issuer_url   = module.eks.cluster_oidc_issuer_url

  discovery_subnet_ids         = module.vpc.private_subnet_ids
  discovery_security_group_ids = [module.eks.cluster_security_group_id]
  tags                         = local.tags

  depends_on = [module.eks]
}

module "finops" {
  source = "../../modules/finops"

  cluster_name               = module.eks.cluster_name
  prometheus_service_name    = "kube-prometheus-stack-prometheus"
  prometheus_namespace       = module.observability.namespace
  monthly_budget_amount      = var.monthly_budget_amount
  budget_notification_emails = var.budget_notification_emails
  tags                       = local.tags

  depends_on = [module.observability]
}

module "backup" {
  source = "../../modules/backup"

  cluster_name      = module.eks.cluster_name
  region            = var.aws_region
  oidc_provider_arn = module.eks.oidc_provider_arn
  oidc_issuer_url   = module.eks.cluster_oidc_issuer_url
  # Short retention in staging; it validates restores, not long-term archival.
  backup_retention_days = 30
  tags                  = local.tags

  depends_on = [module.eks]
}
