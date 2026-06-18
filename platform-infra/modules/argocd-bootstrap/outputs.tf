output "namespace" {
  description = "Argo CD namespace."
  value       = helm_release.argocd.namespace
}

output "release_name" {
  description = "Argo CD Helm release name."
  value       = helm_release.argocd.name
}

