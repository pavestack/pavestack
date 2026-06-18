locals {
  oidc_provider_url = "https://token.actions.githubusercontent.com"
  oidc_provider_arn = var.create_oidc_provider ? aws_iam_openid_connect_provider.github[0].arn : "arn:aws:iam::${data.aws_caller_identity.current.account_id}:oidc-provider/token.actions.githubusercontent.com"
}

data "aws_caller_identity" "current" {}

resource "aws_iam_openid_connect_provider" "github" {
  count = var.create_oidc_provider ? 1 : 0

  url             = local.oidc_provider_url
  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = ["6938fd4d98bab03faadb97b34396831e3780aea1"]

  tags = var.tags
}

data "aws_iam_policy_document" "trust" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]

    principals {
      type        = "Federated"
      identifiers = [local.oidc_provider_arn]
    }

    condition {
      test     = "StringEquals"
      variable = "token.actions.githubusercontent.com:aud"
      values   = ["sts.amazonaws.com"]
    }

    condition {
      test     = "StringLike"
      variable = "token.actions.githubusercontent.com:sub"
      values   = ["repo:${var.github_repository}:environment:${var.github_environment}"]
    }
  }
}

resource "aws_iam_role" "this" {
  name               = "${var.name}-github-actions"
  assume_role_policy = data.aws_iam_policy_document.trust.json

  tags = merge(var.tags, {
    Name = "${var.name}-github-actions"
  })
}

# checkov:skip=CKV_AWS_108: Bootstrap role requires broad access to manage platform infrastructure
data "aws_iam_policy_document" "terraform_bootstrap" {
  statement {
    sid = "TerraformBootstrapServices"
    actions = [
      "autoscaling:*",
      "cloudwatch:*",
      "ec2:*",
      "ecr:*",
      "eks:*",
      "elasticloadbalancing:*",
      "iam:*",
      "kms:*",
      "logs:*",
      "s3:*",
      "ssm:GetParameter",
      "sts:GetCallerIdentity",
      "tag:*"
    ]
    resources = ["*"]
  }
}

resource "aws_iam_policy" "terraform_bootstrap" {
  name        = "${var.name}-terraform-bootstrap"
  description = "Permissions for Pavestack platform-infra Terraform bootstrap."
  policy      = data.aws_iam_policy_document.terraform_bootstrap.json

  tags = var.tags
}

resource "aws_iam_role_policy_attachment" "bootstrap" {
  role       = aws_iam_role.this.name
  policy_arn = aws_iam_policy.terraform_bootstrap.arn
}

resource "aws_iam_role_policy_attachment" "managed" {
  for_each = var.managed_policy_arns

  role       = aws_iam_role.this.name
  policy_arn = each.value
}
