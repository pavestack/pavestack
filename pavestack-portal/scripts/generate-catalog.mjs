import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(__dirname, "..", "..");
const outputPath = path.join(__dirname, "..", "public", "catalog.json");

const searchRoots = [
  path.join(repoRoot, "service-template-api"),
  path.join(repoRoot, "services"),
];

function readField(content, field) {
  const match = content.match(new RegExp(`^\\s*${field}:\\s*(.+)$`, "m"));
  return match ? match[1].trim().replace(/^['"]|['"]$/g, "") : "";
}

function readNested(content, block, field) {
  const blockMatch = content.match(new RegExp(`${block}:\\s*\\n([\\s\\S]*?)(\\n\\S|$)`));
  if (!blockMatch) return "";
  return readField(blockMatch[1], field);
}

function parseScorecard(content) {
  const overall = Number(readField(content, "overall_score") || 0);
  const criteria = [];
  const criteriaBlock = content.match(/criteria:\s*\n([\s\S]*?)(\noverall_score:|$)/);
  if (criteriaBlock) {
    for (const line of criteriaBlock[1].split("\n")) {
      const keyMatch = line.match(/^\s{2}([a-z_]+):\s*$/);
      if (keyMatch) {
        criteria.push({ key: keyMatch[1], status: "unknown", weight: 0 });
      }
      const statusMatch = line.match(/^\s+status:\s*(\S+)/);
      if (statusMatch && criteria.length > 0) {
        criteria[criteria.length - 1].status = statusMatch[1];
      }
      const weightMatch = line.match(/^\s+weight:\s*(\d+)/);
      if (weightMatch && criteria.length > 0) {
        criteria[criteria.length - 1].weight = Number(weightMatch[1]);
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

  const name = readNested(catalogContent, "metadata", "name");
  const description = readNested(catalogContent, "metadata", "description");
  const owner = readNested(catalogContent, "spec", "owner") || readField(scorecardContent, "owner") || "unknown";
  const lifecycle = readNested(catalogContent, "spec", "lifecycle") || "experimental";
  const github = readField(catalogContent, "github.com/project-slug") || "pavestack/pavestack";

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
    scorecard: scorecardContent ? parseScorecard(scorecardContent) : { overall: 0, criteria: [] },
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
