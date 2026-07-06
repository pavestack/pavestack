variable "subscription_id" {
  description = "Azure subscription ID."
  type        = string
  default     = ""
}

variable "location" {
  description = "Azure region."
  type        = string
  default     = "westeurope"
}

variable "environment" {
  description = "Environment name."
  type        = string
  default     = "dev"
}

variable "name_prefix" {
  description = "Resource name prefix."
  type        = string
  default     = "pavestack"
}

variable "resource_group_name" {
  description = "Resource group that owns the environment."
  type        = string
  default     = "pavestack-dev"
}

variable "vnet_cidr" {
  description = "Virtual network CIDR."
  type        = string
  default     = "10.20.0.0/16"
}

variable "kubernetes_version" {
  description = "AKS Kubernetes version."
  type        = string
  default     = "1.30"
}

variable "acr_name" {
  description = "Azure Container Registry name (globally unique, alphanumeric)."
  type        = string
  default     = "pavestackdevacr"
}

variable "node_vm_size" {
  description = "VM size for the default AKS node pool."
  type        = string
  default     = "Standard_D2s_v5"
}

variable "admin_group_object_ids" {
  description = "Entra ID group object IDs granted AKS cluster admin."
  type        = list(string)
  default     = []
}

variable "github_repository" {
  description = "GitHub repository in owner/name form for OIDC trust."
  type        = string
  default     = "pavestack/platform-infra"
}

variable "enable_github_oidc_role" {
  description = "Create a GitHub Actions federated identity for this environment."
  type        = bool
  default     = false
}
