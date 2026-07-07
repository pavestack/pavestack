#!/usr/bin/env node
// scripts/dora-metrics.mjs
//
// DORA-lite metrics for Pavestack. Merged PRs that touch
// platform-config/tenants/** are used as a proxy for "deployment" events,
// since Argo CD applies that tree automatically on merge (see
// docs/metrics.md for the full rationale). No dependencies beyond the
// runtime's built-in `fetch`.
//
// Usage:
//   node scripts/dora-metrics.mjs
//
// Env:
//   GITHUB_REPOSITORY  "owner/repo" (required; GitHub Actions sets this
//                      automatically; set it manually when running locally)
//   GITHUB_TOKEN       optional; raises the API rate limit from 60/hr to
//                      5000/hr and is required for private repos

const WEEKS = 8;
const MS_PER_WEEK = 7 * 24 * 60 * 60 * 1000;
const DEPLOY_PATH_PREFIX = "platform-config/tenants/";
const API = "https://api.github.com";

function repoSlug() {
  const repo = process.env.GITHUB_REPOSITORY;
  if (!repo || !repo.includes("/")) {
    throw new Error(
      "GITHUB_REPOSITORY env var must be set to 'owner/repo' (GitHub Actions " +
        "sets this automatically; set it manually when running locally)."
    );
  }
  return repo;
}

function authHeaders() {
  const headers = {
    Accept: "application/vnd.github+json",
    "User-Agent": "pavestack-dora-metrics",
  };
  if (process.env.GITHUB_TOKEN) {
    headers.Authorization = `Bearer ${process.env.GITHUB_TOKEN}`;
  }
  return headers;
}

async function ghFetch(url) {
  const res = await fetch(url, { headers: authHeaders() });
  if (res.status === 403 || res.status === 429) {
    const remaining = res.headers.get("x-ratelimit-remaining");
    const reset = res.headers.get("x-ratelimit-reset");
    const resetAt = reset ? new Date(Number(reset) * 1000).toISOString() : "unknown";
    throw new Error(
      `GitHub API rate-limited or forbidden (status ${res.status}, ` +
        `remaining=${remaining}, resets at ${resetAt}). ` +
        (process.env.GITHUB_TOKEN
          ? "Check token scopes/permissions."
          : "Set GITHUB_TOKEN to raise the limit (60/hr unauthenticated vs 5000/hr authenticated).")
    );
  }
  if (res.status === 401) {
    throw new Error("GitHub API returned 401 Unauthorized — GITHUB_TOKEN is invalid or expired.");
  }
  if (res.status === 404) {
    throw new Error(`GitHub API returned 404 for ${url} — check GITHUB_REPOSITORY and token access.`);
  }
  if (!res.ok) {
    throw new Error(`GitHub API request failed: ${res.status} ${res.statusText} (${url})`);
  }
  return res.json();
}

// Fetch merged PRs against `base` updated at or after `since`. Closed PRs are
// paginated newest-updated-first, so once a whole page is older than `since`
// there is nothing left worth walking.
async function fetchMergedPRs(slug, since) {
  const prs = [];
  for (let page = 1; page <= 20; page += 1) {
    const url =
      `${API}/repos/${slug}/pulls?state=closed&base=main&per_page=100` +
      `&page=${page}&sort=updated&direction=desc`;
    const batch = await ghFetch(url);
    if (batch.length === 0) break;
    for (const pr of batch) {
      if (pr.merged_at && new Date(pr.merged_at) >= since) prs.push(pr);
    }
    if (batch.every((pr) => new Date(pr.updated_at) < since) || batch.length < 100) break;
  }
  return prs;
}

async function touchesDeployPath(slug, prNumber) {
  for (let page = 1; page <= 10; page += 1) {
    const url = `${API}/repos/${slug}/pulls/${prNumber}/files?per_page=100&page=${page}`;
    const files = await ghFetch(url);
    if (files.some((f) => f.filename.startsWith(DEPLOY_PATH_PREFIX))) return true;
    if (files.length < 100) return false;
  }
  return false;
}

function median(nums) {
  if (nums.length === 0) return null;
  const sorted = [...nums].sort((a, b) => a - b);
  const mid = Math.floor(sorted.length / 2);
  return sorted.length % 2 === 0 ? (sorted[mid - 1] + sorted[mid]) / 2 : sorted[mid];
}

function weekBucket(date, since) {
  return Math.floor((date.getTime() - since.getTime()) / MS_PER_WEEK);
}

async function main() {
  const slug = repoSlug();
  const now = new Date();
  const since = new Date(now.getTime() - WEEKS * MS_PER_WEEK);

  console.log(`DORA-lite metrics for ${slug} — last ${WEEKS} weeks (since ${since.toISOString()})`);
  console.log("Deployment proxy: merged PRs touching platform-config/tenants/** (GitOps apply == deploy)\n");

  const candidates = await fetchMergedPRs(slug, since);
  const deployPRs = [];
  for (const pr of candidates) {
    if (await touchesDeployPath(slug, pr.number)) deployPRs.push(pr);
  }

  if (deployPRs.length === 0) {
    console.log("No merged PRs touching platform-config/tenants/** in the window — nothing to report.");
    return;
  }

  // Deployment frequency: per-week counts.
  const perWeek = new Array(WEEKS).fill(0);
  for (const pr of deployPRs) {
    const bucket = weekBucket(new Date(pr.merged_at), since);
    if (bucket >= 0 && bucket < WEEKS) perWeek[bucket] += 1;
  }
  console.log("Deployment frequency (deploy-proxy merges per week):");
  perWeek.forEach((count, i) => {
    const weekStart = new Date(since.getTime() + i * MS_PER_WEEK).toISOString().slice(0, 10);
    console.log(`  week ${i + 1} (${weekStart}): ${count}`);
  });
  console.log(`  total: ${deployPRs.length}, avg/week: ${(deployPRs.length / WEEKS).toFixed(2)}\n`);

  // Lead time for changes: PR creation -> merge, in hours.
  const leadTimesHours = deployPRs.map(
    (pr) => (new Date(pr.merged_at) - new Date(pr.created_at)) / (1000 * 60 * 60)
  );
  console.log(`Lead time for changes (median, PR open -> merge): ${median(leadTimesHours).toFixed(1)}h\n`);

  // Change failure rate proxy: revert/rollback-titled PRs among the deploy set.
  const failures = deployPRs.filter((pr) => /revert|rollback/i.test(pr.title));
  console.log(
    `Change failure rate proxy: ${failures.length}/${deployPRs.length} deploy PRs titled ` +
      `revert/rollback (${((failures.length / deployPRs.length) * 100).toFixed(1)}%)`
  );
  for (const pr of failures) console.log(`  #${pr.number}: ${pr.title}`);
}

main().catch((err) => {
  console.error(`dora-metrics: ${err.message}`);
  process.exitCode = 1;
});
