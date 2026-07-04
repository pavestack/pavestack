variable "registry_name" {
  description = "Azure Container Registry name (alphanumeric, globally unique)."
  type        = string
}

variable "location" {
  description = "Azure region."
  type        = string
}

variable "resource_group_name" {
  description = "Resource group that owns the registry."
  type        = string
}

variable "retention_days" {
  description = "Number of days to retain untagged manifests."
  type        = number
  default     = 30
}

variable "tags" {
  description = "Tags applied to all resources."
  type        = map(string)
  default     = {}
}
