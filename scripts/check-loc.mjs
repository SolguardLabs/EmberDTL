import { readdirSync, readFileSync, statSync } from "node:fs";
import { join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const root = resolve(fileURLToPath(new URL("..", import.meta.url)));
const src = join(root, "src");
const minimum = 3000;
const maximum = 4000;

function files(dir) {
  const result = [];
  for (const entry of readdirSync(dir)) {
    const full = join(dir, entry);
    const stat = statSync(full);
    if (stat.isDirectory()) {
      result.push(...files(full));
    } else if (full.endsWith(".go")) {
      result.push(full);
    }
  }
  return result;
}

const total = files(src).reduce((sum, file) => {
  const lines = readFileSync(file, "utf8")
    .split(/\r?\n/)
    .filter((line) => line.trim().length > 0).length;
  return sum + lines;
}, 0);

console.log(`src LOC: ${total}`);

if (total < minimum || total > maximum) {
  console.error(`src LOC must stay between ${minimum} and ${maximum}`);
  process.exit(1);
}
