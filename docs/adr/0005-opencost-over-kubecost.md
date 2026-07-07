# 5. OpenCost over Kubecost for cost allocation

Date: 2026-07-01

## Status

Accepted

## Context

Pavestack needs in-cluster cost visibility: attributing spend to namespaces
and tenants so platform and product teams can see what their workloads
cost, feeding the FinOps module (`platform-infra/modules/finops`) alongside
budget alerting (`aws_sns_topic.budget_alerts`). The two well-known options
are Kubecost (a commercial product with a free tier built around its own
bundled Prometheus and UI) and OpenCost (the CNCF sandbox project Kubecost
donated its cost-allocation engine to).

The platform already runs `kube-prometheus-stack` for cluster observability
(`platform-infra/modules/observability`). Kubecost's free tier is easiest
to operate with its own bundled Prometheus, duplicating a component this
cluster already has; its more advanced multi-cluster and long-term
retention features sit behind a paid tier. OpenCost is the underlying
open-source allocation engine (a CNCF project, not tied to one vendor), is
lighter weight (no bundled UI, no bundled second Prometheus), and reads
metrics directly from an existing Prometheus — exactly the shape of this
cluster's observability stack.

## Decision

Deploy OpenCost (`platform-infra/modules/finops/main.tf`, `helm_release
"opencost"` from `https://opencost.github.io/opencost-helm-chart`) and
configure it to read from the observability module's Prometheus rather than
run its own. The module's header comment states this explicitly: "OpenCost
— in-cluster cost allocation, reading usage from the observability module's
Prometheus and attributing spend to namespaces/tenants."

Budget alerting (AWS Budgets → SNS) is handled as a separate,
Terraform-native concern in the same module rather than through OpenCost's
own alerting, since it needs to reach outside the cluster (email/Slack via
SNS) regardless of which in-cluster cost tool is used.

## Consequences

- No second Prometheus to operate or reconcile with the observability
  stack's retention and scrape config — one metrics pipeline, two
  consumers (Grafana dashboards and OpenCost's allocation queries).
- No commercial licensing dependency for cost visibility; OpenCost is
  Apache-2.0 and CNCF-governed, consistent with the rest of the stack's
  preference for community-governed controllers (see ADR 6 for the same
  reasoning applied to policy).
- OpenCost's UI and query surface are less polished than Kubecost's paid
  product; if the platform later needs multi-cluster cost rollups or
  long-term cost retention beyond what this Prometheus retains, that's a
  gap to revisit, not one this ADR closes.
- OpenCost's accuracy depends on the observability module's Prometheus
  staying healthy and retaining enough history — an outage or short
  retention window degrades cost data as well as dashboards.

See also: ADR 3 (why OpenCost, like the rest of the bootstrap-tier
controllers, is Terraform-installed rather than Argo CD-managed).
