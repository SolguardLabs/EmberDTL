import { spawnSync } from "node:child_process";
import { readFileSync, readdirSync, statSync } from "node:fs";
import { extname, join, relative, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const root = resolve(fileURLToPath(new URL("..", import.meta.url)));
const failures = [];

function fail(message) {
  failures.push(message);
}

function walk(directory, extensions) {
  const result = [];
  for (const entry of readdirSync(directory)) {
    const path = join(directory, entry);
    const stat = statSync(path);
    if (stat.isDirectory()) result.push(...walk(path, extensions));
    else if (extensions.has(extname(path))) result.push(path);
  }
  return result;
}

const packageJson = JSON.parse(readFileSync(join(root, "package.json"), "utf8"));
const lock = JSON.parse(readFileSync(join(root, "package-lock.json"), "utf8"));
if (packageJson.version !== "1.0.0" || lock.version !== "1.0.0") {
  fail("package metadata must declare version 1.0.0");
}

const docs = readdirSync(join(root, "docs")).filter((name) => name.endsWith(".md"));
if (docs.length !== 7) fail(`docs must contain exactly 7 Markdown files; found ${docs.length}`);

const markdown = [
  join(root, "README.md"),
  join(root, "SECURITY.md"),
  ...docs.map((name) => join(root, "docs", name)),
];
const markdownText = markdown.map((path) => readFileSync(path, "utf8")).join("\n");
const diagrams = markdownText.match(/^```mermaid$/gm)?.length ?? 0;
if (diagrams !== 27) fail(`expected 27 Mermaid diagrams; found ${diagrams}`);
if (!markdownText.includes("Modelo económico") || !markdownText.includes("reconciliación")) {
  fail("economic model and reconciliation documentation are required");
}
for (const path of markdown) {
  if (readFileSync(path, "utf8").trim().length < 900) {
    fail(`${relative(root, path)} is too short for production documentation`);
  }
}

const publicText = [
  join(root, "README.md"),
  join(root, "SECURITY.md"),
  ...walk(join(root, "docs"), new Set([".md"])),
  ...walk(join(root, "src"), new Set([".go"])),
  ...walk(join(root, "sdk"), new Set([".ts"])),
  ...walk(join(root, "tests"), new Set([".ts", ".json"])),
]
  .map((path) => readFileSync(path, "utf8"))
  .join("\n");
const forbidden =
  /\b(?:ctf|laboratorio|laboratory|vulnerable|vulnerability|exploit|attacker|bug)\b/i;
if (forbidden.test(publicText)) fail("public materials contain a restricted non-operational term");

const banner = readFileSync(join(root, "assets", "banner.png"));
const pngSignature = "89504e470d0a1a0a";
if (banner.subarray(0, 8).toString("hex") !== pngSignature) fail("banner must be a PNG");
const width = banner.readUInt32BE(16);
const height = banner.readUInt32BE(20);
if (width < 1600 || height < 900 || banner.length < 300_000) {
  fail(`banner quality gate failed: ${width}x${height}, ${banner.length} bytes`);
}

const goSources = walk(join(root, "src"), new Set([".go"]));
const nodeTests = walk(join(root, "tests", "node"), new Set([".ts"]));
const goTestCount = goSources.reduce(
  (sum, path) => sum + (readFileSync(path, "utf8").match(/\bfunc Test/g)?.length ?? 0),
  0,
);
const nodeTestCount = nodeTests.reduce(
  (sum, path) => sum + (readFileSync(path, "utf8").match(/\btest\(/g)?.length ?? 0),
  0,
);
if (goTestCount < 10 || nodeTestCount < 10) {
  fail(`test inventory too small: Go=${goTestCount}, Node=${nodeTestCount}`);
}

const protectedFiles = JSON.parse(
  readFileSync(join(root, "scripts", "protected-files.json"), "utf8"),
);
for (const [path, expected] of Object.entries(protectedFiles.files)) {
  const result = spawnSync("git", ["rev-parse", `:${path}`], {
    cwd: root,
    encoding: "utf8",
    shell: false,
  });
  const actual = result.stdout.trim();
  if (result.status !== 0 || actual !== expected) {
    fail(`compatibility lock mismatch for ${path}: ${actual || result.stderr.trim()}`);
  }
}

if (failures.length > 0) {
  for (const failure of failures) console.error(`release gate: ${failure}`);
  process.exit(1);
}

console.log(
  `release gates: ok (docs=${docs.length}, diagrams=${diagrams}, Go tests=${goTestCount}, Node tests=${nodeTestCount})`,
);
