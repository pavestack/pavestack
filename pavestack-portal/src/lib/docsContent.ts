export type DocHeading = {
  id: string;
  title: string;
  body: string[];
};

export type DocSection = {
  id: string;
  title: string;
  headings: DocHeading[];
};

export const DOCS: DocSection[] = [
  {
    id: "architecture",
    title: "Architecture",
    headings: [
      {
        id: "repo-layout",
        title: "Repository layout",
        body: [
          "Pavestack is delivered as a single monorepo. platform-infra/ holds the Terraform modules that provision the VPC, EKS cluster, ECR repositories, and GitHub OIDC trust used by CI. platform-config/ is the GitOps source of truth reconciled continuously by Argo CD — it contains only Kubernetes manifests, Helm values, and Argo CD Application/ApplicationSet definitions, never application source code.",
          "service-template-api/ is the golden-path Go API scaffold that every generated service starts from. pave/ is the self-service CLI (pave create-service) that copies the template, stamps in GitOps manifests, and opens a pull request. pavestack-portal/ (this app) is the developer-facing surface for all of the above. services/ holds the services that have actually been scaffolded via the golden path.",
        ],
      },
      {
        id: "delivery-model",
        title: "Delivery model (GitOps)",
        body: [
          "CI builds artifacts: tests, security scans (Checkov, Trivy, Gitleaks, TFLint), and container images pushed to ECR.",
          "CI opens pull requests: image tags and scaffold changes land in platform-config as PRs rather than being applied directly.",
          "Argo CD reconciles: cluster state follows Git. No kubectl apply is run from a developer laptop or CI job for application workloads — that constraint is enforced by convention across pave, platform-config, and CI.",
        ],
      },
      {
        id: "argocd",
        title: "How Argo CD finds services",
        body: [
          "A root Application is applied once per cluster (manually, by the platform team) and discovers an AppProject plus two ApplicationSets. One ApplicationSet uses a Git directory generator over tenants/*/base to bootstrap each tenant's namespace, RBAC, NetworkPolicy, and ResourceQuota. A second ApplicationSet uses a Git file generator over tenants/*/tenant.yaml to deploy each service's Helm chart into dev and prod.",
          "Because everything is discovered from Git, a new tenant directory under platform-config/tenants/ is sufficient for Argo CD to start managing it after merge — no manual cluster-side registration step is needed beyond the one-time root Application bootstrap.",
        ],
      },
    ],
  },
  {
    id: "onboarding",
    title: "Onboarding",
    headings: [
      {
        id: "prerequisites",
        title: "Prerequisites",
        body: [
          "Go >= 1.23 to build the pave CLI and service-template-api. Node.js >= 20 to build pavestack-portal. Terraform >= 1.6 to manage platform-infra. Docker >= 24 to build container images. The gh CLI (>= 2) is optional but enables pave create-service to open pull requests automatically. AWS CLI (>= 2) is optional for local development.",
        ],
      },
      {
        id: "first-week",
        title: "Your first week",
        body: [
          "1. Clone the monorepo and run `make pave` to build the CLI into ./bin/pave.",
          "2. Scaffold a throwaway service to see the golden path end to end: `pave create-service --name sandbox --team your-team --database=false --no-pr` (the --no-pr flag skips opening a real PR while you're just exploring).",
          "3. Inspect what was generated: a copy of service-template-api under services/sandbox-api, and GitOps manifests under platform-config/tenants/sandbox/.",
          "4. Run this portal locally (`npm ci && npm run dev` inside pavestack-portal/) to see your new service show up in the catalog once catalog.json is regenerated.",
          "5. Run `make test` at the repo root before opening any real pull request.",
        ],
      },
      {
        id: "portal-dev",
        title: "Running the portal locally",
        body: [
          "npm ci && npm run dev inside pavestack-portal/ regenerates catalog.json from the real catalog-info.yaml/scorecard.yaml files in the repo and starts the Vite dev server. Write actions (Create Service, Request Access) call the pave-api backend, which is a separate Go service at pave/cmd/pave-api — set VITE_API_BASE_URL if it isn't running on the default http://localhost:8787/api/v1. If pave-api isn't running, those views show an explicit 'backend unreachable' state rather than a fake success.",
        ],
      },
    ],
  },
  {
    id: "golden-path",
    title: "Golden path explainer",
    headings: [
      {
        id: "what-it-does",
        title: "What `pave create-service` does",
        body: [
          "The CLI is a thin orchestrator: it creates files and opens pull requests, and never runs kubectl apply, helm install, or any command that mutates live cluster state directly. Concretely it: (1) validates the request against schemas/service-request.schema.json, (2) copies service-template-api/ to services/{name}-api and replaces template references (module paths, service name, team owner, Helm chart name), (3) generates GitOps manifests under platform-config/tenants/{name}/ — tenant.yaml, base/kustomization.yaml, and dev/prod application.yaml + values.yaml — and (4) opens a PR via gh pr create unless --no-pr is passed.",
          "Today the template is Go-only. The portal's Create Service wizard reflects that honestly: Node.js and Python are shown as 'coming soon' rather than implying multi-runtime support that doesn't exist yet.",
        ],
      },
      {
        id: "post-scaffold",
        title: "Post-scaffold workflow",
        body: [
          "pave create-service → git commit → git push → PR merge → Argo CD reconciles. Nothing is live in a cluster until the generated manifests are reviewed and merged like any other change.",
        ],
      },
      {
        id: "tenant-onboarding",
        title: "Tenant onboarding details",
        body: [
          "Each tenant directory carries a namespace with pavestack.io/tenant, pavestack.io/team, and pavestack.io/managed-by labels; a default-deny NetworkPolicy for both ingress and egress (services must explicitly allow the traffic they need); least-privilege RBAC granting developers read-only access to their own namespace; and a ResourceQuota capping the namespace at 20 pods, 4 CPU / 8Gi memory requests, and 8 CPU / 16Gi memory limits.",
        ],
      },
    ],
  },
  {
    id: "runbook",
    title: "Runbook template",
    headings: [
      {
        id: "runbook-purpose",
        title: "How to use this template",
        body: [
          "This is a starting template for a per-service on-call runbook, not a record of an existing incident process — copy it into your service's own docs and fill in the specifics. Keep it short enough to read during an active incident.",
        ],
      },
      {
        id: "runbook-sections",
        title: "Suggested sections",
        body: [
          "Service summary: what it does, who owns it (team + Slack channel), its tier, and whether it has a managed database.",
          "Dependencies: upstream/downstream services, external APIs, and the managed database if any.",
          "Dashboards & logs: links to the golden-signals view for this service in Observability, plus wherever structured logs land.",
          "Common failure modes: what typically breaks (e.g. out-of-sync Argo CD Application, failing health checks, exhausted ResourceQuota) and the first three things to check for each.",
          "Rollback: services deploy via Argo CD from Git, so rollback means reverting the image tag change in the tenant's dev/prod values.yaml and letting Argo CD reconcile — not a manual kubectl rollout undo.",
          "Escalation: who to page if the owning team is unreachable.",
        ],
      },
    ],
  },
];
