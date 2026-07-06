variable "name" {
  description = "AKS cluster name."
  type        = string
}

variable "location" {
  description = "Azure region."
  type        = string
}

variable "resource_group_name" {
  description = "Resource group that owns the cluster."
  type        = string
}

variable "kubernetes_version" {
  description = "Kubernetes version for the AKS control plane."
  type        = string
}

variable "subnet_id" {
  description = "Subnet ID for the default node pool."
  type        = string
}

variable "node_vm_size" {
  description = "VM size for the default node pool."
  type        = string
  default     = "Standard_D2s_v5"
}

variable "node_min_count" {
  description = "Minimum default node pool size."
  type        = number
  default     = 1
}

variable "node_max_count" {
  description = "Maximum default node pool size."
  type        = number
  default     = 3
}

variable "acr_id" {
  description = "Container registry resource ID to grant AcrPull to the kubelet identity. Empty to skip."
  type        = string
  default     = ""
}

variable "admin_group_object_ids" {
  description = "Entra ID group object IDs granted cluster admin via Azure RBAC."
  type        = list(string)
  default     = []
}

variable "log_retention_days" {
  description = "Log Analytics workspace retention in days."
  type        = number
  default     = 30
}

variable "tags" {
  description = "Tags applied to all resources."
  type        = map(string)
  default     = {}
}
