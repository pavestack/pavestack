variable "subscription_id" {
  description = "Azure subscription ID."
  type        = string
  default     = ""
}

variable "name_prefix" {
  description = "Resource name prefix."
  type        = string
  default     = "pavestack"
}

variable "location" {
  description = "Azure region for the state storage account."
  type        = string
  default     = "westeurope"
}

variable "resource_group_name" {
  description = "Resource group that holds the Terraform state storage account."
  type        = string
  default     = "pavestack-tfstate"
}

variable "storage_account_name" {
  description = "Globally unique storage account name for Terraform state."
  type        = string
  default     = "pavestacktfstate"
}

variable "container_name" {
  description = "Blob container name for Terraform state."
  type        = string
  default     = "tfstate"
}

variable "retention_days" {
  description = "Soft-delete retention in days for state blobs and containers."
  type        = number
  default     = 30
}

variable "tags" {
  description = "Tags applied to all resources."
  type        = map(string)
  default     = {}
}
