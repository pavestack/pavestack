output "node_role_name" {
  description = "Name of the IAM role attached to nodes launched by Karpenter."
  value       = aws_iam_role.node.name
}

output "node_role_arn" {
  description = "ARN of the IAM role attached to nodes launched by Karpenter."
  value       = aws_iam_role.node.arn
}

output "controller_role_arn" {
  description = "ARN of the IAM role assumed by the Karpenter controller's service account (IRSA)."
  value       = aws_iam_role.controller.arn
}

output "queue_name" {
  description = "Name of the SQS queue Karpenter consumes spot interruption / rebalance / instance state-change / health events from."
  value       = aws_sqs_queue.interruption.name
}

output "namespace" {
  description = "Namespace the Karpenter controller is installed into."
  value       = helm_release.karpenter.namespace
}
