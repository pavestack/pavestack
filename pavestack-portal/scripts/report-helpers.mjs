// Pure helpers for merging committed report artifacts (repo-root `reports/**`) and
// per-service OpenAPI specs into catalog entries. Kept side-effect free (no fs access)
// so they can be unit tested with in-memory fixtures — see report-helpers.test.mjs.
//
// Absence is always tolerated: every function returns `null` when the artifact, or
// the per-service entry inside it, is missing or malformed. The catalog build must
// never fail because a report hasn't been refreshed yet.

const HTTP_METHODS = ["get", "put", "post", "delete", "options", "head", "patch", "trace"];

/**
 * Summarize a Kyverno PolicyReport export (reports/policy/<env>.json) for one service.
 * Returns { pass, warn, fail, passPercent } or null if unknown.
 */
export function computePolicyCompliance(report, serviceName) {
  const entry = report?.namespaces?.[serviceName];
  if (!entry || typeof entry !== "object") return null;

  const pass = Number(entry.pass) || 0;
  const warn = Number(entry.warn) || 0;
  const fail = Number(entry.fail) || 0;
  const total = pass + warn + fail;
  const passPercent = total > 0 ? Math.round((pass / total) * 100) : 0;

  return { pass, warn, fail, passPercent };
}

/**
 * Summarize an OpenCost allocation export (reports/cost/<env>.json) for one service.
 * Returns { amount, currency } or null if unknown.
 */
export function computeCostPerMonth(report, serviceName) {
  const entry = report?.namespaces?.[serviceName];
  if (!entry || typeof entry.monthlyCost !== "number") return null;

  return { amount: entry.monthlyCost, currency: report.currency || "USD" };
}

/**
 * Summarize an Argo CD app list export (reports/deployments/<env>.json) for one service.
 * Returns { syncStatus, health, lastSyncAt } or null if unknown.
 */
export function computeDeploymentHealth(report, serviceName) {
  const entry = report?.apps?.[serviceName];
  if (!entry || typeof entry !== "object") return null;

  return {
    syncStatus: entry.syncStatus || "Unknown",
    health: entry.health || "Unknown",
    lastSyncAt: entry.lastSyncAt || null,
  };
}

/**
 * Parse an already-loaded OpenAPI document (plain object, from YAML or JSON) into the
 * catalog's `api` summary shape. Returns { title, version, endpoints } or null if the
 * document is missing/malformed. Endpoint order follows document order.
 */
export function parseOpenApiDoc(doc) {
  if (!doc || typeof doc !== "object") return null;

  const title = doc.info?.title || "";
  const version = doc.info?.version || "";
  const paths = doc.paths && typeof doc.paths === "object" ? doc.paths : {};

  const endpoints = [];
  for (const [routePath, operations] of Object.entries(paths)) {
    if (!operations || typeof operations !== "object") continue;
    for (const method of HTTP_METHODS) {
      const op = operations[method];
      if (!op || typeof op !== "object") continue;
      endpoints.push({
        method: method.toUpperCase(),
        path: routePath,
        summary: op.summary || op.description || op.operationId || "",
      });
    }
  }

  return { title, version, endpoints };
}
