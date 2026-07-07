output "namespace" {
  description = "Namespace the observability stack is deployed into."
  value       = helm_release.kube_prometheus_stack.namespace
}

output "otel_exporter_otlp_endpoint" {
  description = "Tempo's OTLP/HTTP endpoint. Wire this into services' OTEL_EXPORTER_OTLP_ENDPOINT (e.g. service-template-api)."
  value       = "http://tempo.${var.namespace}.svc.cluster.local:4318"
}

output "otel_exporter_otlp_grpc_endpoint" {
  description = "Tempo's OTLP/gRPC endpoint (host:port, no scheme, for gRPC-based OTLP exporters)."
  value       = "tempo.${var.namespace}.svc.cluster.local:4317"
}

output "prometheus_endpoint" {
  description = "In-cluster Prometheus query endpoint, as exposed by the kube-prometheus-stack chart's prometheus-operator-managed Service."
  value       = "http://kube-prometheus-stack-prometheus.${var.namespace}.svc.cluster.local:9090"
}

output "loki_endpoint" {
  description = "In-cluster Loki endpoint (push API at /loki/api/v1/push, query API at /loki/api/v1/query_range)."
  value       = "http://loki.${var.namespace}.svc.cluster.local:3100"
}

output "grafana_service" {
  description = "Name of the Grafana Kubernetes Service created by the kube-prometheus-stack chart."
  value       = "kube-prometheus-stack-grafana"
}
