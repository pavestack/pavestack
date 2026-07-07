variable "cluster_name" {
  description = "EKS cluster name, reported to OpenCost as the default cluster ID for cost allocation."
  type        = string
}

variable "monthly_budget_amount" {
  description = "Monthly AWS cost budget limit, in USD."
  type        = string
  default     = "100"
}

variable "budget_notification_emails" {
  description = "Email addresses subscribed to AWS Budgets alerts (forecasted and actual overspend)."
  type        = list(string)
  default     = []
}

variable "prometheus_service_name" {
  description = "Kubernetes Service name of the in-cluster Prometheus that OpenCost queries for usage metrics (see the observability module's kube-prometheus-stack release, e.g. kube-prometheus-stack-prometheus)."
  type        = string
}

variable "prometheus_namespace" {
  description = "Namespace the Prometheus Service referenced by prometheus_service_name runs in."
  type        = string
}

variable "prometheus_port" {
  description = "Port the Prometheus Service listens on."
  type        = number
  default     = 9090
}

variable "chart_version" {
  description = "OpenCost Helm chart version."
  type        = string
  default     = "1.40.0"
}

variable "values" {
  description = "Additional Helm values for the OpenCost release."
  type        = list(string)
  default     = []
}

variable "tags" {
  description = "Tags applied to all AWS resources created by this module."
  type        = map(string)
  default     = {}
}
