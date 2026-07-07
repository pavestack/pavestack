# Delivery metrics (DORA-lite)

This page explains the four DORA metrics as Pavestack measures them today,
how to run the measurement script, and where the underlying live data lives.
Pavestack doesn't run a dedicated metrics pipeline (e.g. Four Keys); instead
it derives lightweight proxies from signals the repo and cluster already
produce, good enough to spot trend changes without new infrastructure.

## Why merged tenant PRs are the deployment event

Application and config changes reach `dev`/`staging`/`prod` exclusively
through Argo CD reconciling `platform-config/` (see `docs/promotion.md`).
There is no separate "deploy" step, click, or pipeline run distinct from
merging a change to that tree — merging a PR that edits
`platform-config/tenants/<name>/...` **is** the deploy: Argo CD's
`ApplicationSet` picks up the new commit on `main` and applies it, with
`selfHeal: true` keeping the live state converged on it. That makes "PR
merged, touching `platform-config/tenants/**`" a faithful proxy for a
deployment event, and lets the four metrics below be computed entirely from
the GitHub API with no cluster access required.

## The four metrics

### Deployment frequency

Count of merged `platform-config/tenants/**` PRs per week, over the last 8
weeks. Higher and more even week-to-week is better; a good working
definition of "elite" (per the DORA research) is multiple deploys per day,
but for a platform at Pavestack's current scale, tracking the trend (is
frequency flat, growing, or dropping) matters more than hitting a specific
tier.

### Lead time for changes

Median wall-clock time from PR creation to merge, computed over the same
deploy-proxy PR set. This approximates "commit to production" lead time: it
excludes the time to *author* the change (which varies enormously by task
size) and measures the part the platform controls — review + CI + merge
turnaround for a GitOps change.

### Change failure rate (proxy)

Count of deploy-proxy PRs whose title matches `revert` or `rollback`,
divided by total deploy-proxy PRs in the window. This is a proxy, not a
true failure rate: it only catches failures that were walked back via a
title-flagged revert PR, and it will miss fixes shipped as a normal forward
PR (e.g. "fix: correct replica count") or incidents mitigated without a
config revert at all. Treat it as a floor, not a ceiling, on the true
failure rate.

### Time to restore service

Not currently computed. It requires incident start/end timestamps, which
today live in ad hoc runbook/incident notes (`docs/runbooks/`) rather than a
structured system (e.g. a GitHub issue label with open/close timestamps, or
a PagerDuty/Opsgenie export). If that data becomes structured, this script
is the natural place to add the fourth metric.

## Running the script

```sh
GITHUB_REPOSITORY=pavestack/pavestack GITHUB_TOKEN=<token> node scripts/dora-metrics.mjs
```

- `GITHUB_REPOSITORY` — `owner/repo`; set automatically inside GitHub
  Actions, must be set manually elsewhere.
- `GITHUB_TOKEN` — optional but strongly recommended; unauthenticated
  requests are capped at 60/hour by GitHub, which the per-PR file-list
  lookups can exhaust quickly. A token (even a read-only classic PAT with no
  scopes, for a public repo) raises that to 5000/hour.

The script prints per-week deployment counts, the median lead time in hours,
and the change-failure-rate proxy (with the matching PR numbers/titles) to
stdout. It exits non-zero with a message on missing `GITHUB_REPOSITORY`,
auth failures, and rate-limit errors — there's no silent partial output.

It is not wired into a scheduled workflow yet; run it manually, or add a
scheduled GitHub Actions job that runs it and appends the output to a
`reports/` artifact (see `reports/README.md` for that pattern) if a
standing dashboard is wanted.

## Karpenter compute cost: before/after

Pavestack is migrating cluster autoscaling to Karpenter. Once it has run in
prod for at least two weeks, fill in the table below from the OpenCost
export described in `reports/README.md` (`reports/cost/prod.json`,
before/after snapshots taken two weeks apart around the cutover).

| Metric | Before Karpenter | After Karpenter (2wk) | Delta |
| --- | --- | --- | --- |
| Monthly compute cost (prod) | _to be filled after two weeks of Karpenter data_ | _to be filled after two weeks of Karpenter data_ | _to be filled after two weeks of Karpenter data_ |
| Avg. node utilization | _to be filled after two weeks of Karpenter data_ | _to be filled after two weeks of Karpenter data_ | _to be filled after two weeks of Karpenter data_ |
| Node count (steady state) | _to be filled after two weeks of Karpenter data_ | _to be filled after two weeks of Karpenter data_ | _to be filled after two weeks of Karpenter data_ |

Do not estimate or fabricate these numbers before the data exists — an
empty/placeholder table is more honest than a guessed one.

## Live sources

These metrics and this doc are point-in-time; for current state, use:

- **Grafana dashboards** — cluster and per-tenant resource/cost dashboards
  (ask the platform team for the current URL if not bookmarked; not
  committed to this repo since it's a live service, not a static export).
- **OpenCost UI** — `http://opencost.opencost.svc:9003` in-cluster, or via
  the Grafana OpenCost plugin, for real-time cost allocation (the same API
  `reports/cost/<env>.json` is exported from).
- **`reports/` artifacts** — `reports/cost/`, `reports/deployments/`, and
  `reports/policy/` hold the periodic point-in-time exports consumed by the
  portal catalog; see `reports/README.md` for shape and refresh commands.
- **`scripts/dora-metrics.mjs`** — run on demand for current deployment
  frequency, lead time, and change-failure-rate-proxy numbers.
