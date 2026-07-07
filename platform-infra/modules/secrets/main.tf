locals {
  oidc_issuer_host = replace(var.oidc_issuer_url, "https://", "")

  name               = "${var.cluster_name}-external-secrets"
  namespace          = "external-secrets"
  service_account_sa = "external-secrets"
}

data "aws_caller_identity" "current" {}

# ---------------------------------------------------------------------------
# External Secrets Operator (ESO)
# ---------------------------------------------------------------------------
# ClusterSecretStore and ExternalSecret resources are intentionally NOT
# created here: kubernetes_manifest requires a reachable API server at plan
# time, which breaks `terraform plan` in CI. They are shipped instead as
# kustomize bases at platform-config/templates/{cluster-secret-store,
# external-secret}, applied via Argo CD once the cluster and ESO CRDs exist.

data "aws_iam_policy_document" "external_secrets_assume_role" {
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
      values   = ["system:serviceaccount:${local.namespace}:${local.service_account_sa}"]
    }
  }
}

resource "aws_iam_role" "external_secrets" {
  name               = local.name
  assume_role_policy = data.aws_iam_policy_document.external_secrets_assume_role.json

  tags = merge(var.tags, {
    Name = local.name
  })
}

# Least-privilege read access to the pavestack/<tenant>/<name> Secrets Manager
# path convention. ListSecrets (which cannot be resource-scoped) is
# deliberately omitted: ESO looks up secrets it is explicitly told about via
# ExternalSecret.spec.dataFrom/data, so discovery-style listing isn't needed.
data "aws_iam_policy_document" "external_secrets" {
  statement {
    sid = "ExternalSecretsReadSecrets"
    actions = [
      "secretsmanager:GetSecretValue",
      "secretsmanager:DescribeSecret",
      "secretsmanager:ListSecretVersionIds",
      "secretsmanager:GetResourcePolicy",
    ]
    resources = [
      "arn:aws:secretsmanager:${var.region}:${data.aws_caller_identity.current.account_id}:secret:${var.secret_path_prefix}/*",
    ]
  }
}

resource "aws_iam_policy" "external_secrets" {
  name        = local.name
  description = "Least-privilege Secrets Manager read access for the External Secrets Operator controller on ${var.cluster_name}, scoped to ${var.secret_path_prefix}/*."
  policy      = data.aws_iam_policy_document.external_secrets.json

  tags = var.tags
}

resource "aws_iam_role_policy_attachment" "external_secrets" {
  role       = aws_iam_role.external_secrets.name
  policy_arn = aws_iam_policy.external_secrets.arn
}

resource "helm_release" "external_secrets" {
  name             = "external-secrets"
  repository       = "https://charts.external-secrets.io"
  chart            = "external-secrets"
  version          = var.chart_version
  namespace        = local.namespace
  create_namespace = true
  timeout          = 600
  wait             = true
  atomic           = true

  values = concat([
    yamlencode({
      installCRDs = true
      serviceAccount = {
        create = true
        name   = local.service_account_sa
        annotations = {
          "eks.amazonaws.com/role-arn" = aws_iam_role.external_secrets.arn
        }
      }
    })
  ], var.values)

  depends_on = [aws_iam_role_policy_attachment.external_secrets]
}
