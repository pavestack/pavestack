locals {
  oidc_issuer_host = replace(var.oidc_issuer_url, "https://", "")

  aws_load_balancer_controller_name      = "${var.cluster_name}-aws-load-balancer-controller"
  aws_load_balancer_controller_namespace = "kube-system"
  aws_load_balancer_controller_sa        = "aws-load-balancer-controller"

  external_dns_name      = "${var.cluster_name}-external-dns"
  external_dns_namespace = "external-dns"
  external_dns_sa        = "external-dns"
}

# ---------------------------------------------------------------------------
# AWS Load Balancer Controller
# ---------------------------------------------------------------------------

data "aws_iam_policy_document" "aws_load_balancer_controller_assume_role" {
  count = var.enable_aws_load_balancer_controller ? 1 : 0

  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]

    principals {
      type        = "Federated"
      identifiers = [var.oidc_provider_arn]
    }

    condition {
      test     = "StringEquals"
      variable = "${local.oidc_issuer_host}:aud"
      values   = ["sts.amazonaws.com"]
    }

    condition {
      test     = "StringEquals"
      variable = "${local.oidc_issuer_host}:sub"
      values   = ["system:serviceaccount:${local.aws_load_balancer_controller_namespace}:${local.aws_load_balancer_controller_sa}"]
    }
  }
}

resource "aws_iam_role" "aws_load_balancer_controller" {
  count = var.enable_aws_load_balancer_controller ? 1 : 0

  name               = local.aws_load_balancer_controller_name
  assume_role_policy = data.aws_iam_policy_document.aws_load_balancer_controller_assume_role[0].json

  tags = merge(var.tags, {
    Name = local.aws_load_balancer_controller_name
  })
}

# Official AWS Load Balancer Controller IAM policy (v2.x), reproduced from
# https://raw.githubusercontent.com/kubernetes-sigs/aws-load-balancer-controller/main/docs/install/iam_policy.json
resource "aws_iam_policy" "aws_load_balancer_controller" {
  count = var.enable_aws_load_balancer_controller ? 1 : 0

  name        = local.aws_load_balancer_controller_name
  description = "Permissions for the AWS Load Balancer Controller to manage ALBs/NLBs on behalf of ${var.cluster_name}."
  policy      = file("${path.module}/iam/aws-load-balancer-controller.json")

  tags = var.tags
}

resource "aws_iam_role_policy_attachment" "aws_load_balancer_controller" {
  count = var.enable_aws_load_balancer_controller ? 1 : 0

  role       = aws_iam_role.aws_load_balancer_controller[0].name
  policy_arn = aws_iam_policy.aws_load_balancer_controller[0].arn
}

resource "helm_release" "aws_load_balancer_controller" {
  count = var.enable_aws_load_balancer_controller ? 1 : 0

  name             = "aws-load-balancer-controller"
  repository       = "https://aws.github.io/eks-charts"
  chart            = "aws-load-balancer-controller"
  version          = var.aws_load_balancer_controller_chart_version
  namespace        = local.aws_load_balancer_controller_namespace
  create_namespace = false
  timeout          = 600
  wait             = true
  atomic           = true

  values = concat([
    yamlencode({
      clusterName = var.cluster_name
      region      = var.region
      vpcId       = var.vpc_id
      serviceAccount = {
        create = true
        name   = local.aws_load_balancer_controller_sa
        annotations = {
          "eks.amazonaws.com/role-arn" = aws_iam_role.aws_load_balancer_controller[0].arn
        }
      }
    })
  ], var.aws_load_balancer_controller_values)

  depends_on = [aws_iam_role_policy_attachment.aws_load_balancer_controller]
}

# ---------------------------------------------------------------------------
# cert-manager
# ---------------------------------------------------------------------------
# ClusterIssuer resources are intentionally NOT created here: kubernetes_manifest
# requires a reachable API server at plan time, which breaks `terraform plan` in
# CI. The ClusterIssuer is shipped instead as a kustomize base at
# platform-config/templates/cluster-issuer, applied via Argo CD once the cluster
# and cert-manager CRDs exist.

resource "helm_release" "cert_manager" {
  count = var.enable_cert_manager ? 1 : 0

  name             = "cert-manager"
  repository       = "https://charts.jetstack.io"
  chart            = "cert-manager"
  version          = var.cert_manager_chart_version
  namespace        = "cert-manager"
  create_namespace = true
  timeout          = 600
  wait             = true
  atomic           = true

  values = concat([
    yamlencode({
      crds = {
        enabled = true
      }
    })
  ], var.cert_manager_values)
}

# ---------------------------------------------------------------------------
# external-dns
# ---------------------------------------------------------------------------

data "aws_iam_policy_document" "external_dns_assume_role" {
  count = var.enable_external_dns ? 1 : 0

  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]

    principals {
      type        = "Federated"
      identifiers = [var.oidc_provider_arn]
    }

    condition {
      test     = "StringEquals"
      variable = "${local.oidc_issuer_host}:aud"
      values   = ["sts.amazonaws.com"]
    }

    condition {
      test     = "StringEquals"
      variable = "${local.oidc_issuer_host}:sub"
      values   = ["system:serviceaccount:${local.external_dns_namespace}:${local.external_dns_sa}"]
    }
  }
}

data "aws_iam_policy_document" "external_dns" {
  count = var.enable_external_dns ? 1 : 0

  statement {
    sid       = "ExternalDNSChangeRecordSets"
    actions   = ["route53:ChangeResourceRecordSets"]
    resources = ["arn:aws:route53:::hostedzone/${var.route53_zone_id}"]
  }

  # AWS does not support resource-level scoping for these read/list actions.
  statement {
    sid = "ExternalDNSListZones"
    actions = [
      "route53:ListHostedZones",
      "route53:ListResourceRecordSets",
      "route53:ListTagsForResource",
    ]
    resources = ["*"]
  }
}

resource "aws_iam_role" "external_dns" {
  count = var.enable_external_dns ? 1 : 0

  name               = local.external_dns_name
  assume_role_policy = data.aws_iam_policy_document.external_dns_assume_role[0].json

  tags = merge(var.tags, {
    Name = local.external_dns_name
  })
}

resource "aws_iam_policy" "external_dns" {
  count = var.enable_external_dns ? 1 : 0

  name        = local.external_dns_name
  description = "Least-privilege Route53 permissions for external-dns on ${var.cluster_name}, scoped to hosted zone ${var.route53_zone_id}."
  policy      = data.aws_iam_policy_document.external_dns[0].json

  tags = var.tags
}

resource "aws_iam_role_policy_attachment" "external_dns" {
  count = var.enable_external_dns ? 1 : 0

  role       = aws_iam_role.external_dns[0].name
  policy_arn = aws_iam_policy.external_dns[0].arn
}

resource "helm_release" "external_dns" {
  count = var.enable_external_dns ? 1 : 0

  name             = "external-dns"
  repository       = "https://kubernetes-sigs.github.io/external-dns"
  chart            = "external-dns"
  version          = var.external_dns_chart_version
  namespace        = local.external_dns_namespace
  create_namespace = true
  timeout          = 600
  wait             = true
  atomic           = true

  values = concat([
    yamlencode({
      provider = {
        name = "aws"
      }
      aws = {
        region = var.region
      }
      domainFilters = [var.domain_filter]
      policy        = var.external_dns_policy
      txtOwnerId    = var.cluster_name
      serviceAccount = {
        create = true
        name   = local.external_dns_sa
        annotations = {
          "eks.amazonaws.com/role-arn" = aws_iam_role.external_dns[0].arn
        }
      }
    })
  ], var.external_dns_values)

  depends_on = [aws_iam_role_policy_attachment.external_dns]
}
