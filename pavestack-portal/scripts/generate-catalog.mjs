import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import * as yaml from "js-yaml";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(__dirname, "..", "..");
const outputPath = path.join(__dirname, "..", "public", "catalog.json");
const tenantsRoot = path.join(repoRoot, "platform-config", "tenants");

const searchRoots = [path.join(repoRoot, "service-template-api"), path.join(repoRoot, "services")];

const KNOWN_TIERS = new Set(["tier-1", "tier-2", "tier-3"]);
const KNOWN_EXPOSURES = new Set(["internal", "public"]);

function toLabel(key) {
  return key.replaceAll("_", " ").replace(/\b\w/g, (c) => c.toUpperCase());
}

function parseScorecard(scorecardObj) {
  if (!scorecardObj) {
    return { overallScore: 0, criteria: [] };
  }
  const overallScore =
    typeof scorecardObj.overall_score === "number"
      ? scorecardObj.overall_score
      : scorecardObj.overall_score
        ? Number(scorecardObj.overall_score)
        : 0;

  const criteria = [];
  if (scorecardObj.criteria && typeof scorecardObj.criteria === "object") {
    for (const [key, val] of Object.entries(scorecardObj.criteria)) {
      if (val && typeof val === "object") {
        criteria.push({
          key,
          label: toLabel(key),
          status: val.status || "unknown",
          weight: typeof val.weight === "number" ? val.weight : val.weight ? Number(val.weight) : 0,
          evidence: val.evidence || null,
        });
      } else {
        criteria.push({ key, label: toLabel(key), status: "unknown", weight: 0, evidence: null });
      }
    }
  }
  return { overallScore, criteria };
}

function walkServices(root) {
  const services = [];
  if (!fs.existsSync(root)) return services;

  for (const entry of fs.readdirSync(root, { withFileTypes: true })) {
    if (!entry.isDirectory()) continue;
    const fullPath = path.join(root, entry.name);
    if (fs.existsSync(path.join(fullPath, "catalog-info.yaml"))) {
      services.push(fullPath);
    }
  }
  return services;
}

function collectServiceDirs(root) {
  if (!fs.existsSync(root)) return [];
  if (fs.existsSync(path.join(root, "catalog-info.yaml"))) {
    return [root];
  }
  return walkServices(root);
}

/**
 * The tenant directory name (platform-config/tenants/<name>) is the bare
 * `--name` value passed to `pave create-service`, while the service source
 * directory is `services/<name>-api`. Try the exact match first, then strip
 * a trailing `-api` suffix, so tenant-derived data (image tags, sync state)
 * still resolves for generated services.
 */
function findTenantDir(serviceName) {
  const candidates = [serviceName, serviceName.replace(/-api$/, "")];
  for (const candidate of candidates) {
    const candidatePath = path.join(tenantsRoot, candidate);
    if (fs.existsSync(candidatePath)) return candidatePath;
  }
  return null;
}

const EMPTY_TENANT_META = { database: null, tier: null, runtime: null, exposure: null };

/**
 * tenant.yaml is written by `pave create-service` and is the most authoritative
 * source for tier/runtime/exposure/database — preferred over catalog-info.yaml
 * annotations when present.
 */
function readTenantMeta(tenantDir) {
  if (!tenantDir) return EMPTY_TENANT_META;
  const tenantYamlPath = path.join(tenantDir, "tenant.yaml");
  if (!fs.existsSync(tenantYamlPath)) return EMPTY_TENANT_META;
  try {
    const tenant = yaml.load(fs.readFileSync(tenantYamlPath, "utf8")) || {};
    return {
      database: typeof tenant.database === "boolean" ? tenant.database : null,
      tier: KNOWN_TIERS.has(tenant.tier) ? tenant.tier : null,
      runtime: typeof tenant.runtime === "string" ? tenant.runtime : null,
      exposure: KNOWN_EXPOSURES.has(tenant.exposure) ? tenant.exposure : null,
    };
  } catch {
    return EMPTY_TENANT_META;
  }
}

function readEnvValues(tenantDir, env) {
  if (!tenantDir) return {};
  const valuesPath = path.join(tenantDir, env, "values.yaml");
  if (!fs.existsSync(valuesPath)) return {};
  try {
    return yaml.load(fs.readFileSync(valuesPath, "utf8")) || {};
  } catch {
    return {};
  }
}

function readImageTag(values) {
  return values.image?.tag ?? null;
}

/** Configured CPU/memory requests+limits from committed Helm values — real allocation, not live usage. */
function readResources(values) {
  const resources = values.resources;
  if (!resources || typeof resources !== "object") return null;
  const pick = (block) =>
    block && typeof block === "object"
      ? { cpu: block.cpu ?? null, memory: block.memory ?? null }
      : null;
  const requests = pick(resources.requests);
  const limits = pick(resources.limits);
  if (!requests && !limits) return null;
  return { requests, limits };
}

function loadService(dir) {
  const catalogContent = fs.readFileSync(path.join(dir, "catalog-info.yaml"), "utf8");
  const scorecardPath = path.join(dir, "scorecard.yaml");
  const scorecardContent = fs.existsSync(scorecardPath)
    ? fs.readFileSync(scorecardPath, "utf8")
    : "";

  const catalog = yaml.load(catalogContent) || {};
  const scorecardObj = scorecardContent ? yaml.load(scorecardContent) || {} : null;

  const name = catalog.metadata?.name || "";
  const description = catalog.metadata?.description || "";
  const annotations = catalog.metadata?.annotations || {};
  const owner = catalog.spec?.owner || scorecardObj?.owner || "unknown";
  const lifecycle = catalog.spec?.lifecycle || "experimental";
  const system = catalog.spec?.system || "pavestack";
  const github = annotations["github.com/project-slug"] || "pavestack/pavestack";

  // Stamped by `pave create-service` into generated catalog-info.yaml files.
  // Absence means the service/catalog entry was authored by hand.
  const createdVia = annotations["pavestack.io/created-via"] === "pave-cli" ? "pave-cli" : "manual";

  const tenantDir = findTenantDir(name);
  const tenantMeta = readTenantMeta(tenantDir);

  // tenant.yaml (written by pave create-service) is preferred; fall back to
  // catalog-info.yaml annotations for services onboarded before that field existed.
  const tier =
    tenantMeta.tier ??
    (KNOWN_TIERS.has(annotations["pavestack.io/tier"]) ? annotations["pavestack.io/tier"] : null);
  const runtime = tenantMeta.runtime ?? annotations["pavestack.io/runtime"] ?? null;
  const exposure =
    tenantMeta.exposure ??
    (KNOWN_EXPOSURES.has(annotations["pavestack.io/exposure"])
      ? annotations["pavestack.io/exposure"]
      : null);

  const devValues = readEnvValues(tenantDir, "dev");
  const prodValues = readEnvValues(tenantDir, "prod");

  return {
    id: name,
    name,
    description,
    // `owner` and `team` currently map to the same `spec.owner` slug — the
    // service-request schema only tracks one owning-team concept today.
    owner,
    team: owner,
    system,
    lifecycle,
    tier,
    runtime,
    exposure,
    database: tenantMeta.database,
    createdVia,
    repoUrl: `https://github.com/${github}`,
    repoPath: path.relative(repoRoot, dir),
    // NOTE: status/health below are illustrative placeholders, not a live
    // Argo CD/Prometheus read. imageTag and resources (when resolvable) are
    // real, sourced from the tenant's committed Helm values — resources are
    // the *configured allocation*, not live usage.
    environments: {
      dev: {
        status: "synced",
        health: "healthy",
        imageTag: readImageTag(devValues),
        resources: readResources(devValues),
      },
      prod: {
        status: "synced",
        health: "healthy",
        imageTag: readImageTag(prodValues),
        resources: readResources(prodValues),
      },
    },
    scorecard: parseScorecard(scorecardObj),
  };
}

const serviceDirs = searchRoots.flatMap(collectServiceDirs);
const catalog = {
  generatedAt: new Date().toISOString(),
  services: serviceDirs.map(loadService).sort((a, b) => a.name.localeCompare(b.name)),
};

fs.mkdirSync(path.dirname(outputPath), { recursive: true });
fs.writeFileSync(outputPath, JSON.stringify(catalog, null, 2));
console.log(`Wrote ${catalog.services.length} services to ${outputPath}`);
