import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import yaml from "js-yaml";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(__dirname, "..", "..");
const outputPath = path.join(__dirname, "..", "public", "catalog.json");

const searchRoots = [
  path.join(repoRoot, "service-template-api"),
  path.join(repoRoot, "services"),
];

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

  const catalog = yaml.load(catalogContent) || {};
  const scorecardObj = scorecardContent ? (yaml.load(scorecardContent) || {}) : null;

  const name = catalog.metadata?.name || "";
  const description = catalog.metadata?.description || "";
  const owner = catalog.spec?.owner || scorecardObj?.owner || "unknown";
  const lifecycle = catalog.spec?.lifecycle || "experimental";
  const github = catalog.metadata?.annotations?.["github.com/project-slug"] || "pavestack/pavestack";

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

