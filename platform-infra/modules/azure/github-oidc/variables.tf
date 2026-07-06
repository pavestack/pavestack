variable "name" {
  description = "Identity name prefix."
  type        = string
}

variable "location" {
  description = "Azure region."
  type        = string
}

variable "resource_group_name" {
  description = "Resource group that owns the identity."
  type        = string
}

variable "github_repository" {
  description = "GitHub repository in owner/name form."
  type        = string
}

variable "github_environment" {
  description = "GitHub environment allowed to federate with the identity."
  type        = string
}

variable "role_assignments" {
  description = "Map of role assignments (role definition name => scope) granted to the identity."
  type        = map(string)
  default     = {}
}

variable "tags" {
  description = "Tags applied to all resources."
  type        = map(string)
  default     = {}
}
