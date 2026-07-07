locals {
  # Alertmanager routes everything to a null (blackhole) receiver by default so
  # the module works standalone. When var.alert_webhook_url is set, a second
  # "slack" receiver is added and becomes the default route's receiver. The
  # Watchdog heartbeat alert is always routed to "null" regardless, matching
  # kube-prometheus-stack's own default (it's a liveness signal, not something
  # that should page anyone).
  alertmanager_default_receiver = var.alert_webhook_url != "" ? "slack" : "null"

  alertmanager_receivers = concat(
    [{ name = "null" }],
    var.alert_webhook_url != "" ? [{
      name = "slack"
      slack_configs = [{
        api_url       = var.alert_webhook_url
        send_resolved = true
        channel       = "#platform-alerts"
        title         = "{{ .CommonAnnotations.summary }}"
        text          = "{{ .CommonAnnotations.description }}"
      }]
    }] : []
  )
}

# ---------------------------------------------------------------------------
# kube-prometheus-stack: Prometheus, Grafana, Alertmanager, prometheus-operator
# ---------------------------------------------------------------------------

resource "helm_release" "kube_prometheus_stack" {
  name             = "kube-prometheus-stack"
  repository       = "https://prometheus-community.github.io/helm-charts"
  chart            = "kube-prometheus-stack"
  version          = var.kube_prometheus_stack_chart_version
  namespace        = var.namespace
  create_namespace = true
  timeout          = 1800
  wait             = true
  atomic           = true

  values = concat([
    yamlencode({
      defaultRules = {
        create = true
      }

      prometheus = {
        prometheusSpec = {
          retention                               = var.prometheus_retention
          serviceMonitorSelectorNilUsesHelmValues = false
          podMonitorSelectorNilUsesHelmValues     = false
          storageSpec = {
            volumeClaimTemplate = {
              spec = {
                accessModes = ["ReadWriteOnce"]
                resources = {
                  requests = {
                    storage = var.prometheus_storage_size
                  }
                }
              }
            }
          }
        }
      }

      alertmanager = {
        config = {
          route = {
            receiver        = local.alertmanager_default_receiver
            group_by        = ["namespace", "alertname"]
            group_wait      = "30s"
            group_interval  = "5m"
            repeat_interval = "4h"
            routes = [
              {
                receiver = "null"
                matchers = ["alertname = \"Watchdog\""]
              }
            ]
          }
          receivers = local.alertmanager_receivers
        }
      }

      # Argo CD sync/health rules + node pressure aren't covered by
      # defaultRules (the kubernetes-mixin rules ship KubePodCrashLooping,
      # KubeNodeNotReady, etc., but nothing about Argo CD app state or
      # node condition pressure), so they're added explicitly here.
      additionalPrometheusRulesMap = {
        platform-observability-rules = {
          groups = [
            {
              name = "argocd.rules"
              rules = [
                {
                  alert = "ArgoCDAppSyncFailed"
                  expr  = "argocd_app_info{sync_status=\"OutOfSync\"} == 1"
                  for   = "15m"
                  labels = {
                    severity = "warning"
                  }
                  annotations = {
                    summary     = "Argo CD application {{ $labels.name }} is OutOfSync"
                    description = "Application {{ $labels.name }} in project {{ $labels.project }} has been OutOfSync for more than 15 minutes."
                  }
                },
                {
                  alert = "ArgoCDAppHealthDegraded"
                  expr  = "argocd_app_info{health_status=\"Degraded\"} == 1"
                  for   = "10m"
                  labels = {
                    severity = "critical"
                  }
                  annotations = {
                    summary     = "Argo CD application {{ $labels.name }} health is Degraded"
                    description = "Application {{ $labels.name }} in project {{ $labels.project }} has reported Degraded health for more than 10 minutes."
                  }
                }
              ]
            },
            {
              name = "node-pressure.rules"
              rules = [
                {
                  alert = "NodeUnderResourcePressure"
                  expr  = "max by (node) (kube_node_status_condition{condition=~\"MemoryPressure|DiskPressure|PIDPressure\", status=\"true\"}) == 1"
                  for   = "5m"
                  labels = {
                    severity = "warning"
                  }
                  annotations = {
                    summary     = "Node {{ $labels.node }} is under resource pressure"
                    description = "Node {{ $labels.node }} has reported a MemoryPressure, DiskPressure or PIDPressure condition for more than 5 minutes."
                  }
                }
              ]
            }
          ]
        }
      }

      grafana = {
        additionalDataSources = [
          {
            name      = "Loki"
            type      = "loki"
            uid       = "loki"
            access    = "proxy"
            url       = "http://loki.${var.namespace}.svc.cluster.local:3100"
            isDefault = false
            jsonData = {
              maxLines = 1000
            }
          },
          {
            name      = "Tempo"
            type      = "tempo"
            uid       = "tempo"
            access    = "proxy"
            url       = "http://tempo.${var.namespace}.svc.cluster.local:3100"
            isDefault = false
            jsonData = {
              httpMethod = "GET"
              tracesToLogsV2 = {
                datasourceUid      = "loki"
                spanStartTimeShift = "-1h"
                spanEndTimeShift   = "1h"
                filterByTraceID    = false
                filterBySpanID     = false
              }
              serviceMap = {
                datasourceUid = "prometheus"
              }
            }
          }
        ]

        dashboardProviders = {
          "dashboardproviders.yaml" = {
            apiVersion = 1
            providers = [
              {
                name            = "default"
                orgId           = 1
                folder          = ""
                type            = "file"
                disableDeletion = false
                editable        = true
                options = {
                  path = "/var/lib/grafana/dashboards/default"
                }
              }
            ]
          }
        }

        dashboards = {
          default = {
            cluster-overview = {
              gnetId     = 15757
              revision   = 1
              datasource = "Prometheus"
            }
            node-exporter-full = {
              gnetId     = 1860
              revision   = 37
              datasource = "Prometheus"
            }
            namespace-pod-resources = {
              gnetId     = 15758
              revision   = 1
              datasource = "Prometheus"
            }
            golden-signals = {
              gnetId     = 15761
              revision   = 1
              datasource = "Prometheus"
            }
          }
        }
      }
    })
  ], var.kube_prometheus_stack_values)
}

# ---------------------------------------------------------------------------
# Loki: single-binary / filesystem mode, suitable for in-cluster demo use
# ---------------------------------------------------------------------------

resource "helm_release" "loki" {
  name             = "loki"
  repository       = "https://grafana.github.io/helm-charts"
  chart            = "loki"
  version          = var.loki_chart_version
  namespace        = var.namespace
  create_namespace = false
  timeout          = 1200
  wait             = true
  atomic           = true

  values = concat([
    yamlencode({
      loki = {
        auth_enabled = false
        commonConfig = {
          replication_factor = 1
        }
        schemaConfig = {
          configs = [
            {
              from         = "2024-01-01"
              store        = "tsdb"
              object_store = "filesystem"
              schema       = "v13"
              index = {
                prefix = "loki_index_"
                period = "24h"
              }
            }
          ]
        }
        storage = {
          type = "filesystem"
        }
        limits_config = {
          retention_period = "${var.loki_retention_hours}h"
        }
        compactor = {
          retention_enabled    = true
          delete_request_store = "filesystem"
        }
      }

      singleBinary = {
        replicas = 1
        persistence = {
          enabled = true
          size    = "10Gi"
        }
      }

      read = {
        replicas = 0
      }
      write = {
        replicas = 0
      }
      backend = {
        replicas = 0
      }

      minio = {
        enabled = false
      }

      gateway = {
        enabled = false
      }

      test = {
        enabled = false
      }

      monitoring = {
        selfMonitoring = {
          enabled = false
        }
        lokiCanary = {
          enabled = false
        }
        serviceMonitor = {
          enabled = false
        }
      }
    })
  ], var.loki_values)

  depends_on = [helm_release.kube_prometheus_stack]
}

# ---------------------------------------------------------------------------
# Promtail: ships pod logs from every node to Loki
# ---------------------------------------------------------------------------

resource "helm_release" "promtail" {
  name             = "promtail"
  repository       = "https://grafana.github.io/helm-charts"
  chart            = "promtail"
  version          = var.promtail_chart_version
  namespace        = var.namespace
  create_namespace = false
  timeout          = 1200
  wait             = true
  atomic           = true

  values = concat([
    yamlencode({
      config = {
        clients = [
          {
            url = "http://loki.${var.namespace}.svc.cluster.local:3100/loki/api/v1/push"
          }
        ]
      }
    })
  ], var.promtail_values)

  depends_on = [helm_release.loki]
}

# ---------------------------------------------------------------------------
# Tempo: monolithic mode, OTLP gRPC (4317) / HTTP (4318) receivers
# ---------------------------------------------------------------------------

resource "helm_release" "tempo" {
  name             = "tempo"
  repository       = "https://grafana.github.io/helm-charts"
  chart            = "tempo"
  version          = var.tempo_chart_version
  namespace        = var.namespace
  create_namespace = false
  timeout          = 1200
  wait             = true
  atomic           = true

  values = concat([
    yamlencode({
      tempo = {
        receivers = {
          otlp = {
            protocols = {
              grpc = {
                endpoint = "0.0.0.0:4317"
              }
              http = {
                endpoint = "0.0.0.0:4318"
              }
            }
          }
        }
      }

      persistence = {
        enabled = true
        size    = "10Gi"
      }
    })
  ], var.tempo_values)

  depends_on = [helm_release.kube_prometheus_stack]
}
