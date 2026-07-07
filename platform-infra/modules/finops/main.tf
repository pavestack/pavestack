data "aws_caller_identity" "current" {}

# ---------------------------------------------------------------------------
# OpenCost — in-cluster cost allocation, reading usage from the observability
# module's Prometheus and attributing spend to namespaces/tenants.
# ---------------------------------------------------------------------------

resource "helm_release" "opencost" {
  name             = "opencost"
  repository       = "https://opencost.github.io/opencost-helm-chart"
  chart            = "opencost"
  version          = var.chart_version
  namespace        = "opencost"
  create_namespace = true
  timeout          = 600
  wait             = true
  atomic           = true

  values = concat([
    yamlencode({
      opencost = {
        prometheus = {
          internal = {
            serviceName   = var.prometheus_service_name
            namespaceName = var.prometheus_namespace
            port          = var.prometheus_port
          }
        }
        exporter = {
          defaultClusterId = var.cluster_name
        }
        ui = {
          enabled = true
        }
      }
    })
  ], var.values)
}

# ---------------------------------------------------------------------------
# AWS Budgets — monthly cost budget with forecasted and actual overspend
# alerts published to SNS.
#
# Cost-allocation *tag activation* (Billing and Cost Management > Cost
# allocation tags) is a payer/management-account operation and is not, and
# cannot be, managed from this member-account module. The
# pavestack.io/tenant and pavestack.io/cost-center namespace labels applied
# via platform-config/templates/namespace give OpenCost and any activated
# AWS cost-allocation tags a consistent key to join on once that activation
# happens out-of-band.
# ---------------------------------------------------------------------------

# Customer-managed key so budget-alert encryption is auditable and revocable
# (the AWS-managed alias/aws/sns key allows no key-policy control). Budgets
# needs kms:GenerateDataKey*/Decrypt to publish to the encrypted topic.
resource "aws_kms_key" "budget_alerts" {
  description             = "KMS key for ${var.cluster_name} budget alert SNS encryption"
  deletion_window_in_days = 10
  enable_key_rotation     = true
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "Enable IAM User Permissions"
        Effect    = "Allow"
        Principal = { AWS = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root" }
        Action    = "kms:*"
        Resource  = "*"
      },
      {
        Sid       = "AllowBudgetsUseOfKey"
        Effect    = "Allow"
        Principal = { Service = "budgets.amazonaws.com" }
        Action    = ["kms:GenerateDataKey*", "kms:Decrypt"]
        Resource  = "*"
        Condition = {
          StringEquals = { "aws:SourceAccount" = data.aws_caller_identity.current.account_id }
        }
      }
    ]
  })

  tags = merge(var.tags, {
    Name = "${var.cluster_name}-budget-alerts"
  })
}

resource "aws_kms_alias" "budget_alerts" {
  name          = "alias/${var.cluster_name}-budget-alerts"
  target_key_id = aws_kms_key.budget_alerts.key_id
}

resource "aws_sns_topic" "budget_alerts" {
  name              = "${var.cluster_name}-budget-alerts"
  kms_master_key_id = aws_kms_key.budget_alerts.arn

  tags = merge(var.tags, {
    Name = "${var.cluster_name}-budget-alerts"
  })
}

data "aws_iam_policy_document" "budget_alerts_publish" {
  statement {
    sid     = "AllowBudgetsPublish"
    effect  = "Allow"
    actions = ["SNS:Publish"]

    principals {
      type        = "Service"
      identifiers = ["budgets.amazonaws.com"]
    }

    resources = [aws_sns_topic.budget_alerts.arn]

    condition {
      test     = "StringEquals"
      variable = "aws:SourceAccount"
      values   = [data.aws_caller_identity.current.account_id]
    }
  }
}

resource "aws_sns_topic_policy" "budget_alerts" {
  arn    = aws_sns_topic.budget_alerts.arn
  policy = data.aws_iam_policy_document.budget_alerts_publish.json
}

resource "aws_sns_topic_subscription" "budget_alerts_email" {
  for_each = toset(var.budget_notification_emails)

  topic_arn = aws_sns_topic.budget_alerts.arn
  protocol  = "email"
  endpoint  = each.value
}

resource "aws_budgets_budget" "monthly" {
  name         = "${var.cluster_name}-monthly-cost"
  budget_type  = "COST"
  limit_amount = var.monthly_budget_amount
  limit_unit   = "USD"
  time_unit    = "MONTHLY"

  notification {
    comparison_operator       = "GREATER_THAN"
    threshold                 = 80
    threshold_type            = "PERCENTAGE"
    notification_type         = "FORECASTED"
    subscriber_sns_topic_arns = [aws_sns_topic.budget_alerts.arn]
  }

  notification {
    comparison_operator       = "GREATER_THAN"
    threshold                 = 100
    threshold_type            = "PERCENTAGE"
    notification_type         = "ACTUAL"
    subscriber_sns_topic_arns = [aws_sns_topic.budget_alerts.arn]
  }

  depends_on = [aws_sns_topic_policy.budget_alerts]
}
