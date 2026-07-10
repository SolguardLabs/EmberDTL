import { existsSync, mkdirSync, rmSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";

const root = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const binDir = join(root, "bin");
const exeName = process.platform === "win32" ? "emberdtl.exe" : "emberdtl";
const output = join(binDir, exeName);
const args = new Set(process.argv.slice(2));

function run(command, commandArgs, options = {}) {
  return spawnSync(command, commandArgs, {
    cwd: root,
    encoding: "utf8",
    stdio: options.stdio ?? "pipe",
    shell: false,
  });
}

function commandExists(command) {
  if (process.platform === "win32") {
    return run("where.exe", [command]).status === 0;
  }
  return (
    spawnSync("sh", ["-c", `command -v ${command}`], {
      cwd: root,
      encoding: "utf8",
      stdio: "pipe",
    }).status === 0
  );
}

if (args.has("--clean")) {
  rmSync(binDir, { recursive: true, force: true });
}

if (!commandExists("go")) {
  console.error("Go toolchain not found. Install Go 1.22+ or make sure go is available on PATH.");
  process.exit(1);
}

mkdirSync(binDir, { recursive: true });

const build = run("go", ["build", "-o", output, "./src/cmd/emberdtl"], { stdio: "inherit" });
if (build.status !== 0) {
  process.exit(build.status ?? 1);
}

if (!existsSync(output)) {
  console.error(`Build did not create ${output}`);
  process.exit(1);
}

console.log(output);
