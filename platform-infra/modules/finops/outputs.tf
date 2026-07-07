output "namespace" {
  description = "Namespace OpenCost is deployed into."
  value       = helm_release.opencost.namespace
}

output "sns_topic_arn" {
  description = "ARN of the SNS topic AWS Budgets notifications are published to."
  value       = aws_sns_topic.budget_alerts.arn
}

output "budget_name" {
  description = "Name of the AWS Budgets monthly cost budget."
  value       = aws_budgets_budget.monthly.name
}
