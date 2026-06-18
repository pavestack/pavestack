variable "namespace" {
  description = "Namespace for Argo CD."
  type        = string
  default     = "argocd"
}

variable "chart_version" {
  description = "Argo CD Helm chart version."
  type        = string
}

variable "values" {
  description = "Additional Helm values for Argo CD."
  type        = list(string)
  default     = []
}

