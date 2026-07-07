resource "helm_release" "argocd" {
  name             = "argocd"
  repository       = "https://argoproj.github.io/argo-helm"
  chart            = "argo-cd"
  version          = var.chart_version
  namespace        = var.namespace
  create_namespace = true
  timeout          = 1200
  wait             = true
  atomic           = true

  values = concat([
    yamlencode({
      global = {
        domain = null
      }
      server = {
        service = {
          type = "ClusterIP"
        }
        metrics = {
          enabled = true
          serviceMonitor = {
            enabled = true
          }
        }
      }
      # Expose Prometheus metrics + ServiceMonitors so the observability module's
      # Argo CD alert rules (which query argocd_app_info) have something to scrape.
      controller = {
        metrics = {
          enabled = true
          serviceMonitor = {
            enabled = true
          }
        }
      }
      repoServer = {
        metrics = {
          enabled = true
          serviceMonitor = {
            enabled = true
          }
        }
      }
      configs = {
        params = {
          "server.insecure" = false
        }
      }
    })
  ], var.values)
}

