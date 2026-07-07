variable "namespace" {
  description = "Namespace for Kyverno."
  type        = string
  default     = "kyverno"
}

variable "chart_version" {
  description = "Kyverno Helm chart version."
  type        = string
}

variable "values" {
  description = "Additional Helm values for Kyverno (e.g. override admissionController.replicas for HA in prod)."
  type        = list(string)
  default     = []
}
