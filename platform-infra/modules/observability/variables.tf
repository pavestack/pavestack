variable "namespace" {
  description = "Namespace for the observability stack (Prometheus, Grafana, Alertmanager, Loki, Promtail, Tempo)."
  type        = string
  default     = "observability"
}

variable "kube_prometheus_stack_chart_version" {
  description = "kube-prometheus-stack Helm chart version (Prometheus, Grafana, Alertmanager, prometheus-operator)."
  type        = string
  default     = "65.5.1"
}

variable "loki_chart_version" {
  description = "Loki Helm chart version (single-binary/filesystem mode)."
  type        = string
  default     = "6.6.4"
}

variable "promtail_chart_version" {
  description = "Promtail Helm chart version (log shipping agent for Loki)."
  type        = string
  default     = "6.16.6"
}

variable "tempo_chart_version" {
  description = "Tempo Helm chart version (monolithic mode, OTLP receivers)."
  type        = string
  default     = "1.10.1"
}

variable "alert_webhook_url" {
  description = "Slack-compatible incoming webhook URL for Alertmanager notifications. When empty, alerts are routed to a null/blackhole receiver and the module works without any external dependency."
  type        = string
  default     = ""
  sensitive   = true
}

variable "prometheus_retention" {
  description = "How long Prometheus retains time-series data."
  type        = string
  default     = "15d"
}

variable "prometheus_storage_size" {
  description = "Size of the persistent volume claim for Prometheus's storage."
  type        = string
  default     = "20Gi"
}

variable "loki_retention_hours" {
  description = "Number of hours Loki retains log data before compaction deletes it."
  type        = number
  default     = 168
}

variable "kube_prometheus_stack_values" {
  description = "Additional Helm values for the kube-prometheus-stack release, applied after the module's defaults."
  type        = list(string)
  default     = []
}

variable "loki_values" {
  description = "Additional Helm values for the Loki release, applied after the module's defaults."
  type        = list(string)
  default     = []
}

variable "promtail_values" {
  description = "Additional Helm values for the Promtail release, applied after the module's defaults."
  type        = list(string)
  default     = []
}

variable "tempo_values" {
  description = "Additional Helm values for the Tempo release, applied after the module's defaults."
  type        = list(string)
  default     = []
}
