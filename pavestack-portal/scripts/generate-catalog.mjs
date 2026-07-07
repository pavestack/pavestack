import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { load as parseYaml } from "js-yaml";
import {
  computePolicyCompliance,
  computeCostPerMonth,
  computeDeploymentHealth,
  parseOpenApiDoc,
} from "./report-helpers.mjs";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(__dirname, "..", "..");
const outputPath = path.join(__dirname, "..", "public", "catalog.json");
const reportsRoot = path.join(repoRoot, "reports");

// Environments the platform provisions for every tenant (see platform-config/tenants).
const ENVS = ["dev", "prod"];

const searchRoots = [
  path.join(repoRoot, "service-template-api"),
  path.join(repoRoot, "services"),
];

// reports/<kind>/<env>.json are read once per (kind, env) and shared across services.
const reportCache = new Map();

function readJsonIfPresent(filePath) {
  if (!fs.existsSync(filePath)) return null;
  try {
    return JSON.parse(fs.readFileSync(filePath, "utf8"));
  } catch (err) {
    console.warn(`generate-catalog: failed to parse ${path.relative(repoRoot, filePath)}: ${err.message}`);
    return null;
  }
}

function getReport(kind, env) {
  const key = `${kind}/${env}`;
  if (!reportCache.has(key)) {
    reportCache.set(key, readJsonIfPresent(path.join(reportsRoot, kind, `${env}.json`)));
  }
  return reportCache.get(key);
}

function loadApiSummary(dir) {
  const openApiPath = path.join(dir, "openapi.yaml");
  if (!fs.existsSync(openApiPath)) return null;
  try {
    const doc = parseYaml(fs.readFileSync(openApiPath, "utf8"));
    return parseOpenApiDoc(doc);
  } catch (err) {
    console.warn(`generate-catalog: failed to parse ${path.relative(repoRoot, openApiPath)}: ${err.message}`);
    return null;
  }
}

function parseScorecard(scorecardObj) {
  if (!scorecardObj) {
    return { overall: 0, criteria: [] };
  }
  const overall = typeof scorecardObj.overall_score === "number"
    ? scorecardObj.overall_score
    : (scorecardObj.overall_score ? Number(scorecardObj.overall_score) : 0);

  const criteria = [];
  if (scorecardObj.criteria && typeof scorecardObj.criteria === "object") {
    for (const [key, val] of Object.entries(scorecardObj.criteria)) {
      if (val && typeof val === "object") {
        criteria.push({
          key,
          status: val.status || "unknown",
          weight: typeof val.weight === "number"
            ? val.weight
            : (val.weight ? Number(val.weight) : 0),
        });
      } else {
        criteria.push({
          key,
          status: "unknown",
          weight: 0,
        });
      }
    }
  }
  return { overall, criteria };
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

function loadService(dir) {
  const catalogContent = fs.readFileSync(path.join(dir, "catalog-info.yaml"), "utf8");
  const scorecardPath = path.join(dir, "scorecard.yaml");
  const scorecardContent = fs.existsSync(scorecardPath)
    ? fs.readFileSync(scorecardPath, "utf8")
    : "";

  const catalog = parseYaml(catalogContent) || {};
  const scorecardObj = scorecardContent ? (parseYaml(scorecardContent) || {}) : null;

  const name = catalog.metadata?.name || "";
  const description = catalog.metadata?.description || "";
  const owner = catalog.spec?.owner || scorecardObj?.owner || "unknown";
  const lifecycle = catalog.spec?.lifecycle || "experimental";
  const github = catalog.metadata?.annotations?.["github.com/project-slug"] || "pavestack/pavestack";

  // Report-derived signals, one entry per environment. Any of these is `null` when
  // the corresponding reports/<kind>/<env>.json is absent, or doesn't mention this
  // service yet — the UI renders that as "unknown" rather than failing the build.
  const policyCompliance = {};
  const costPerMonth = {};
  const deploymentHealth = {};
  for (const env of ENVS) {
    policyCompliance[env] = computePolicyCompliance(getReport("policy", env), name);
    costPerMonth[env] = computeCostPerMonth(getReport("cost", env), name);
    deploymentHealth[env] = computeDeploymentHealth(getReport("deployments", env), name);
  }

  return {
    id: name,
    name,
    description,
    owner,
    repo: `https://github.com/${github}`,
    repoPath: path.relative(repoRoot, dir),
    lifecycle,
    environments: {
      dev: { status: "synced", health: "healthy" },
      prod: { status: "synced", health: "healthy" },
    },
    scorecard: parseScorecard(scorecardObj),
    policyCompliance,
    costPerMonth,
    deploymentHealth,
    api: loadApiSummary(dir),
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

