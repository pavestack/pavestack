locals {
  account_id        = data.aws_caller_identity.current.account_id
  oidc_provider_url = "https://token.actions.githubusercontent.com"
  oidc_provider_arn = var.create_oidc_provider ? aws_iam_openid_connect_provider.github[0].arn : "arn:aws:iam::${local.account_id}:oidc-provider/token.actions.githubusercontent.com"

  # bootstrap/remote-state names the bucket <name_prefix>-<environment>-tfstate-<account_id>,
  # and var.name is <name_prefix>-<environment>.
  state_bucket_arns = length(var.state_bucket_arns) > 0 ? var.state_bucket_arns : ["arn:aws:s3:::${var.name}-tfstate-*"]
  state_object_arns = [for arn in local.state_bucket_arns : "${arn}/*"]

  secret_arns = [for prefix in var.secret_name_prefixes : "arn:aws:secretsmanager:*:${local.account_id}:secret:${prefix}*"]
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

# Terraform plan/apply permissions for the whole platform, one statement per service,
# resource-scoped to the ${var.name} prefix wherever the action supports it. Split across
# two managed policies to stay under the 6144-character policy size limit:
#   - terraform_bootstrap: state backend plus EC2/EKS/ECR/logs/SSM infrastructure.
#   - terraform_bootstrap_iam: IAM/KMS/CloudWatch/budgets/Secrets Manager/Route53, plus the
#     read-only and wildcard-exception statements.
data "aws_iam_policy_document" "terraform_bootstrap" {
  # Terraform remote state: read/write state objects and S3 lockfiles in the tfstate bucket(s).
  statement {
    sid = "TerraformStateObjects"
    actions = [
      "s3:GetObject",
      "s3:PutObject",
      "s3:DeleteObject"
    ]
    resources = local.state_object_arns
  }

  statement {
    sid = "TerraformStateBucket"
    actions = [
      "s3:ListBucket",
      "s3:GetBucketVersioning",
      "s3:GetBucketLocation"
    ]
    resources = local.state_bucket_arns
  }

  # Optional DynamoDB state locking; empty by default because the backend uses S3 lockfiles.
  dynamic "statement" {
    for_each = length(var.lock_table_arns) > 0 ? [1] : []

    content {
      sid = "TerraformStateLock"
      actions = [
        "dynamodb:DescribeTable",
        "dynamodb:GetItem",
        "dynamodb:PutItem",
        "dynamodb:DeleteItem"
      ]
      resources = var.lock_table_arns
    }
  }

  # State objects are SSE-KMS encrypted; allow use of account keys only via S3 so this
  # statement cannot be used to decrypt anything directly.
  statement {
    sid = "StateKmsViaS3"
    actions = [
      "kms:Encrypt",
      "kms:Decrypt",
      "kms:ReEncryptFrom",
      "kms:ReEncryptTo",
      "kms:GenerateDataKey",
      "kms:GenerateDataKeyWithoutPlaintext"
    ]
    resources = ["arn:aws:kms:*:${local.account_id}:key/*"]

    condition {
      test     = "StringLike"
      variable = "kms:ViaService"
      values   = ["s3.*.amazonaws.com"]
    }
  }

  # VPC module: VPC, subnets, routing, IGW/NAT/EIP, security groups, flow logs, tags.
  # EC2 mutating actions accept resource ARNs but the IDs are unknowable before apply,
  # so scope to this account's EC2 resources (blocks cross-account resource access).
  statement {
    sid = "Ec2NetworkWrite"
    actions = [
      "ec2:CreateVpc",
      "ec2:DeleteVpc",
      "ec2:ModifyVpcAttribute",
      "ec2:CreateSubnet",
      "ec2:DeleteSubnet",
      "ec2:ModifySubnetAttribute",
      "ec2:CreateRouteTable",
      "ec2:DeleteRouteTable",
      "ec2:CreateRoute",
      "ec2:DeleteRoute",
      "ec2:ReplaceRoute",
      "ec2:AssociateRouteTable",
      "ec2:DisassociateRouteTable",
      "ec2:ReplaceRouteTableAssociation",
      "ec2:CreateInternetGateway",
      "ec2:DeleteInternetGateway",
      "ec2:AttachInternetGateway",
      "ec2:DetachInternetGateway",
      "ec2:AllocateAddress",
      "ec2:ReleaseAddress",
      "ec2:DisassociateAddress",
      "ec2:CreateNatGateway",
      "ec2:DeleteNatGateway",
      "ec2:CreateSecurityGroup",
      "ec2:DeleteSecurityGroup",
      "ec2:AuthorizeSecurityGroupIngress",
      "ec2:AuthorizeSecurityGroupEgress",
      "ec2:RevokeSecurityGroupIngress",
      "ec2:RevokeSecurityGroupEgress",
      "ec2:ModifySecurityGroupRules",
      "ec2:UpdateSecurityGroupRuleDescriptionsIngress",
      "ec2:UpdateSecurityGroupRuleDescriptionsEgress",
      "ec2:CreateFlowLogs",
      "ec2:DeleteFlowLogs",
      "ec2:CreateTags",
      "ec2:DeleteTags"
    ]
    resources = ["arn:aws:ec2:*:${local.account_id}:*"]
  }

  # EKS module: cluster, node groups, addons, access entries — all named from ${var.name}.
  # ASGs and launch templates for managed node groups are created by the EKS service-linked
  # role, so no autoscaling/ec2-instance write access is needed here.
  statement {
    sid = "EksManage"
    actions = [
      "eks:*"
    ]
    resources = [
      "arn:aws:eks:*:${local.account_id}:cluster/${var.name}*",
      "arn:aws:eks:*:${local.account_id}:nodegroup/${var.name}*/*",
      "arn:aws:eks:*:${local.account_id}:addon/${var.name}*/*",
      "arn:aws:eks:*:${local.account_id}:access-entry/${var.name}*/*",
      "arn:aws:eks:*:${local.account_id}:podidentityassociation/${var.name}*/*",
      "arn:aws:eks:*:${local.account_id}:identityproviderconfig/${var.name}*/*"
    ]
  }

  # ECR module plus CI image pushes: manage and push to platform repositories only.
  statement {
    sid = "EcrManageRepositories"
    actions = [
      "ecr:CreateRepository",
      "ecr:DeleteRepository",
      "ecr:DescribeRepositories",
      "ecr:TagResource",
      "ecr:UntagResource",
      "ecr:ListTagsForResource",
      "ecr:PutLifecyclePolicy",
      "ecr:GetLifecyclePolicy",
      "ecr:DeleteLifecyclePolicy",
      "ecr:PutImageScanningConfiguration",
      "ecr:PutImageTagMutability",
      "ecr:GetRepositoryPolicy",
      "ecr:SetRepositoryPolicy",
      "ecr:DeleteRepositoryPolicy",
      "ecr:DescribeImages",
      "ecr:ListImages",
      "ecr:BatchGetImage",
      "ecr:BatchCheckLayerAvailability",
      "ecr:GetDownloadUrlForLayer",
      "ecr:InitiateLayerUpload",
      "ecr:UploadLayerPart",
      "ecr:CompleteLayerUpload",
      "ecr:PutImage",
      "ecr:BatchDeleteImage"
    ]
    resources = ["arn:aws:ecr:*:${local.account_id}:repository/${var.ecr_repository_prefix}/*"]
  }

  # CloudWatch log groups for the EKS control plane and VPC flow logs.
  statement {
    sid = "LogsManageGroups"
    actions = [
      "logs:CreateLogGroup",
      "logs:DeleteLogGroup",
      "logs:PutRetentionPolicy",
      "logs:DeleteRetentionPolicy",
      "logs:AssociateKmsKey",
      "logs:DisassociateKmsKey",
      "logs:TagResource",
      "logs:UntagResource",
      "logs:ListTagsForResource"
    ]
    resources = [
      "arn:aws:logs:*:${local.account_id}:log-group:/aws/eks/${var.name}*",
      "arn:aws:logs:*:${local.account_id}:log-group:/aws/vpc-flow-logs/${var.name}*"
    ]
  }

  # Public SSM parameters published by AWS (EKS optimized AMI/addon metadata lookups).
  statement {
    sid = "SsmReadPublicParameters"
    actions = [
      "ssm:GetParameter",
      "ssm:GetParameters",
      "ssm:GetParametersByPath"
    ]
    resources = ["arn:aws:ssm:*::parameter/aws/service/*"]
  }
}

# Companion document: IAM/KMS/alerting/DNS/secrets, the *-only read actions, and the
# documented wildcard exceptions.
data "aws_iam_policy_document" "terraform_bootstrap_iam" {
  # checkov:skip=CKV_AWS_107:ecr:GetAuthorizationToken does not support resource-level scoping; required for docker login when CI pushes images
  # checkov:skip=CKV_AWS_111:kms:CreateKey does not support resource-level scoping; keys for EKS secrets and VPC flow logs are created by Terraform

  # IAM roles/policies created by the platform modules (cluster, node, ebs-csi, flow logs,
  # ALB controller, external-dns, this module itself) — all named with the ${var.name} prefix.
  statement {
    sid = "IamRolesAndPolicies"
    actions = [
      "iam:CreateRole",
      "iam:DeleteRole",
      "iam:UpdateRole",
      "iam:UpdateRoleDescription",
      "iam:UpdateAssumeRolePolicy",
      "iam:TagRole",
      "iam:UntagRole",
      "iam:PutRolePolicy",
      "iam:DeleteRolePolicy",
      "iam:AttachRolePolicy",
      "iam:DetachRolePolicy",
      "iam:CreatePolicy",
      "iam:DeletePolicy",
      "iam:CreatePolicyVersion",
      "iam:DeletePolicyVersion",
      "iam:SetDefaultPolicyVersion",
      "iam:TagPolicy",
      "iam:UntagPolicy"
    ]
    resources = [
      "arn:aws:iam::${local.account_id}:role/${var.name}*",
      "arn:aws:iam::${local.account_id}:policy/${var.name}*"
    ]
  }

  # OIDC providers: this module's GitHub provider and the EKS cluster IRSA provider (whose
  # issuer host is generated by EKS, so it cannot be name-scoped).
  statement {
    sid = "IamOidcProviders"
    actions = [
      "iam:CreateOpenIDConnectProvider",
      "iam:DeleteOpenIDConnectProvider",
      "iam:TagOpenIDConnectProvider",
      "iam:UntagOpenIDConnectProvider",
      "iam:UpdateOpenIDConnectProviderThumbprint",
      "iam:AddClientIDToOpenIDConnectProvider",
      "iam:RemoveClientIDFromOpenIDConnectProvider"
    ]
    resources = ["arn:aws:iam::${local.account_id}:oidc-provider/*"]
  }

  # PassRole is limited to platform-prefixed roles and only towards the services that
  # actually consume them (EKS cluster/node roles, VPC flow log delivery role).
  statement {
    sid       = "IamPassRoleToServices"
    actions   = ["iam:PassRole"]
    resources = ["arn:aws:iam::${local.account_id}:role/${var.name}*"]

    condition {
      test     = "StringEquals"
      variable = "iam:PassedToService"
      values = [
        "eks.amazonaws.com",
        "vpc-flow-logs.amazonaws.com"
      ]
    }
  }

  # Service-linked roles EKS and ELB create on first use.
  statement {
    sid       = "IamServiceLinkedRoles"
    actions   = ["iam:CreateServiceLinkedRole"]
    resources = ["arn:aws:iam::${local.account_id}:role/aws-service-role/*"]

    condition {
      test     = "StringEquals"
      variable = "iam:AWSServiceName"
      values = [
        "eks.amazonaws.com",
        "eks-nodegroup.amazonaws.com",
        "elasticloadbalancing.amazonaws.com",
        "autoscaling.amazonaws.com"
      ]
    }
  }

  # KMS key lifecycle for platform keys (EKS secrets, VPC flow logs, tfstate, ECR grants).
  # Key IDs are random UUIDs so key ARNs cannot be name-scoped; aliases can.
  statement {
    sid = "KmsManageKeys"
    actions = [
      "kms:DescribeKey",
      "kms:GetKeyPolicy",
      "kms:GetKeyRotationStatus",
      "kms:ListResourceTags",
      "kms:PutKeyPolicy",
      "kms:EnableKeyRotation",
      "kms:DisableKeyRotation",
      "kms:ScheduleKeyDeletion",
      "kms:CancelKeyDeletion",
      "kms:TagResource",
      "kms:UntagResource",
      "kms:UpdateKeyDescription",
      "kms:CreateGrant",
      "kms:ListGrants",
      "kms:RevokeGrant",
      "kms:RetireGrant",
      "kms:CreateAlias",
      "kms:DeleteAlias",
      "kms:UpdateAlias"
    ]
    resources = [
      "arn:aws:kms:*:${local.account_id}:key/*",
      "arn:aws:kms:*:${local.account_id}:alias/${var.name}*"
    ]
  }

  # CloudWatch alarms, scoped to the platform prefix (platform alerting).
  statement {
    sid = "CloudWatchAlarms"
    actions = [
      "cloudwatch:PutMetricAlarm",
      "cloudwatch:DeleteAlarms",
      "cloudwatch:TagResource",
      "cloudwatch:UntagResource",
      "cloudwatch:ListTagsForResource"
    ]
    resources = ["arn:aws:cloudwatch:*:${local.account_id}:alarm:${var.name}*"]
  }

  # Cost budgets, scoped to the platform prefix (budgets ARNs carry no region).
  statement {
    sid = "BudgetsManage"
    actions = [
      "budgets:ViewBudget",
      "budgets:ModifyBudget"
    ]
    resources = ["arn:aws:budgets::${local.account_id}:budget/${var.name}*"]
  }

  # Read-only access to platform secrets so Terraform can plan External Secrets Operator
  # resources; runtime secret reads use the ESO IRSA role, not this one.
  statement {
    sid = "SecretsManagerRead"
    actions = [
      "secretsmanager:DescribeSecret",
      "secretsmanager:GetSecretValue",
      "secretsmanager:GetResourcePolicy",
      "secretsmanager:ListSecretVersionIds"
    ]
    resources = local.secret_arns
  }

  # Route53 record writes only in explicitly allowed zones; empty by default because
  # external-dns owns record management at runtime with its own zone-scoped role.
  dynamic "statement" {
    for_each = length(var.route53_zone_arns) > 0 ? [1] : []

    content {
      sid = "Route53ManageRecords"
      actions = [
        "route53:ChangeResourceRecordSets",
        "route53:ChangeTagsForResource"
      ]
      resources = var.route53_zone_arns
    }
  }

  # Read-only describe/list actions that AWS only supports with Resource "*". Terraform
  # needs these to refresh state and resolve data sources. autoscaling/elasticloadbalancing
  # are read-only here on purpose: ASGs are managed by the EKS service-linked role and load
  # balancers by the ALB controller's IRSA role.
  statement {
    sid = "ReadOnly"
    actions = [
      "autoscaling:Describe*",
      "cloudwatch:Describe*",
      "ec2:Describe*",
      "eks:DescribeAddonVersions",
      "eks:DescribeAddonConfiguration",
      "eks:ListClusters",
      "eks:ListAccessPolicies",
      "elasticloadbalancing:Describe*",
      "iam:Get*",
      "iam:List*",
      "kms:ListAliases",
      "kms:ListKeys",
      "logs:Describe*",
      "route53:Get*",
      "route53:List*",
      "secretsmanager:ListSecrets",
      "sts:GetCallerIdentity",
      "tag:Get*"
    ]
    resources = ["*"]
  }

  # Explicit wildcard exceptions — the only non-read actions granted on "*", because AWS
  # does not support resource-level scoping for them:
  #   - ecr:GetAuthorizationToken: registry-level docker login, required by CI image pushes.
  #   - kms:CreateKey: keys have no ARN until created; management of the resulting keys is
  #     constrained by KmsManageKeys above.
  statement {
    sid = "WildcardOnlyActions"
    actions = [
      "ecr:GetAuthorizationToken",
      "kms:CreateKey"
    ]
    resources = ["*"]
  }
}

resource "aws_iam_policy" "terraform_bootstrap" {
  name        = "${var.name}-terraform-bootstrap"
  description = "Pavestack platform-infra Terraform: state backend and EC2/EKS/ECR/logs infrastructure."
  policy      = data.aws_iam_policy_document.terraform_bootstrap.json

  tags = var.tags
}

resource "aws_iam_policy" "terraform_bootstrap_iam" {
  name        = "${var.name}-terraform-bootstrap-iam"
  description = "Pavestack platform-infra Terraform: IAM, KMS, alerting, DNS, secrets and read-only access."
  policy      = data.aws_iam_policy_document.terraform_bootstrap_iam.json

  tags = var.tags
}

resource "aws_iam_role_policy_attachment" "bootstrap" {
  role       = aws_iam_role.this.name
  policy_arn = aws_iam_policy.terraform_bootstrap.arn
}

resource "aws_iam_role_policy_attachment" "bootstrap_iam" {
  role       = aws_iam_role.this.name
  policy_arn = aws_iam_policy.terraform_bootstrap_iam.arn
}

resource "aws_iam_role_policy_attachment" "managed" {
  for_each = var.managed_policy_arns

  role       = aws_iam_role.this.name
  policy_arn = each.value
}
