# observability module

Deploys the platform's observability stack: kube-prometheus-stack (Prometheus, Grafana, Alertmanager), Loki (single-binary/filesystem mode) with Promtail for log shipping, and Tempo (monolithic mode) for traces, wired together with Grafana datasources and a starter dashboard/alert-rule set. Alertmanager routes to a null receiver unless a Slack webhook URL is supplied.

<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.6 |
| <a name="requirement_helm"></a> [helm](#requirement\_helm) | >= 2.17 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_helm"></a> [helm](#provider\_helm) | >= 2.17 |

## Resources

| Name | Type |
|------|------|
| helm_release.kube_prometheus_stack | resource |
| helm_release.loki | resource |
| helm_release.promtail | resource |
| helm_release.tempo | resource |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_alert_webhook_url"></a> [alert\_webhook\_url](#input\_alert\_webhook\_url) | Slack-compatible incoming webhook URL for Alertmanager notifications. When empty, alerts are routed to a null/blackhole receiver and the module works without any external dependency. | `string` | `""` | no |
| <a name="input_kube_prometheus_stack_chart_version"></a> [kube\_prometheus\_stack\_chart\_version](#input\_kube\_prometheus\_stack\_chart\_version) | kube-prometheus-stack Helm chart version (Prometheus, Grafana, Alertmanager, prometheus-operator). | `string` | `"65.5.1"` | no |
| <a name="input_kube_prometheus_stack_values"></a> [kube\_prometheus\_stack\_values](#input\_kube\_prometheus\_stack\_values) | Additional Helm values for the kube-prometheus-stack release, applied after the module's defaults. | `list(string)` | `[]` | no |
| <a name="input_loki_chart_version"></a> [loki\_chart\_version](#input\_loki\_chart\_version) | Loki Helm chart version (single-binary/filesystem mode). | `string` | `"6.6.4"` | no |
| <a name="input_loki_retention_hours"></a> [loki\_retention\_hours](#input\_loki\_retention\_hours) | Number of hours Loki retains log data before compaction deletes it. | `number` | `168` | no |
| <a name="input_loki_values"></a> [loki\_values](#input\_loki\_values) | Additional Helm values for the Loki release, applied after the module's defaults. | `list(string)` | `[]` | no |
| <a name="input_namespace"></a> [namespace](#input\_namespace) | Namespace for the observability stack (Prometheus, Grafana, Alertmanager, Loki, Promtail, Tempo). | `string` | `"observability"` | no |
| <a name="input_prometheus_retention"></a> [prometheus\_retention](#input\_prometheus\_retention) | How long Prometheus retains time-series data. | `string` | `"15d"` | no |
| <a name="input_prometheus_storage_size"></a> [prometheus\_storage\_size](#input\_prometheus\_storage\_size) | Size of the persistent volume claim for Prometheus's storage. | `string` | `"20Gi"` | no |
| <a name="input_promtail_chart_version"></a> [promtail\_chart\_version](#input\_promtail\_chart\_version) | Promtail Helm chart version (log shipping agent for Loki). | `string` | `"6.16.6"` | no |
| <a name="input_promtail_values"></a> [promtail\_values](#input\_promtail\_values) | Additional Helm values for the Promtail release, applied after the module's defaults. | `list(string)` | `[]` | no |
| <a name="input_tempo_chart_version"></a> [tempo\_chart\_version](#input\_tempo\_chart\_version) | Tempo Helm chart version (monolithic mode, OTLP receivers). | `string` | `"1.10.1"` | no |
| <a name="input_tempo_values"></a> [tempo\_values](#input\_tempo\_values) | Additional Helm values for the Tempo release, applied after the module's defaults. | `list(string)` | `[]` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_grafana_service"></a> [grafana\_service](#output\_grafana\_service) | Name of the Grafana Kubernetes Service created by the kube-prometheus-stack chart. |
| <a name="output_loki_endpoint"></a> [loki\_endpoint](#output\_loki\_endpoint) | In-cluster Loki endpoint (push API at /loki/api/v1/push, query API at /loki/api/v1/query\_range). |
| <a name="output_namespace"></a> [namespace](#output\_namespace) | Namespace the observability stack is deployed into. |
| <a name="output_otel_exporter_otlp_endpoint"></a> [otel\_exporter\_otlp\_endpoint](#output\_otel\_exporter\_otlp\_endpoint) | Tempo's OTLP/HTTP endpoint. Wire this into services' OTEL\_EXPORTER\_OTLP\_ENDPOINT (e.g. service-template-api). |
| <a name="output_otel_exporter_otlp_grpc_endpoint"></a> [otel\_exporter\_otlp\_grpc\_endpoint](#output\_otel\_exporter\_otlp\_grpc\_endpoint) | Tempo's OTLP/gRPC endpoint (host:port, no scheme, for gRPC-based OTLP exporters). |
| <a name="output_prometheus_endpoint"></a> [prometheus\_endpoint](#output\_prometheus\_endpoint) | In-cluster Prometheus query endpoint, as exposed by the kube-prometheus-stack chart's prometheus-operator-managed Service. |
<!-- END_TF_DOCS -->
