# Kyverno admission controller — validates the baseline ClusterPolicy set
# delivered via GitOps (platform-config/policies). No AWS IAM is required:
# Kyverno only talks to the Kubernetes API server it runs inside.
resource "helm_release" "kyverno" {
  name             = "kyverno"
  repository       = "https://kyverno.github.io/kyverno"
  chart            = "kyverno"
  version          = var.chart_version
  namespace        = var.namespace
  create_namespace = true
  timeout          = 900
  wait             = true
  atomic           = true

  values = concat([
    yamlencode({
      admissionController = {
        # Single replica is enough for the dev demo. Prod should override this
        # to 3 for HA admission webhooks, e.g.:
        #   values = [yamlencode({ admissionController = { replicas = 3 } })]
        replicas = 1
      }
    })
  ], var.values)
}
