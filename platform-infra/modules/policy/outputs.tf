output "namespace" {
  description = "Kyverno namespace."
  value       = helm_release.kyverno.namespace
}

output "release_name" {
  description = "Kyverno Helm release name."
  value       = helm_release.kyverno.name
}
