import { spawnSync } from "node:child_process";
import { existsSync, statSync } from "node:fs";
import { isAbsolute, resolve } from "node:path";

export type AmountEntry = { id: string; amount: number };

export type EmberReport = {
  name: string;
  epoch: number;
  assets: Array<{ id: string; symbol: string; decimals: number }>;
  accounts: Array<{
    id: string;
    role: string;
    status: string;
    free: AmountEntry[];
    held: AmountEntry[];
    claimsPaid: AmountEntry[];
  }>;
  pools: Array<{
    id: string;
    asset: string;
    status: string;
    balance: number;
    contributions: number;
    claimsPaid: number;
    pendingDefaults: number;
    pendingClaims: number;
  }>;
  reserves: Array<Record<string, number | string>>;
  facilities: Array<Record<string, number | string>>;
  defaults: Array<Record<string, number | string | string[]>>;
  claims: Array<Record<string, number | string>>;
  reconciliation: Array<{
    asset: string;
    reserveBalance: number;
    facilityExposure: number;
    poolBalance: number;
    pendingDefaults: number;
    pendingClaims: number;
    netCoverageBuffer: number;
  }>;
  metrics: Record<string, number>;
  validations?: string[];
};

export type ClientOptions = {
  binaryPath: string;
  cwd?: string;
  timeoutMs?: number;
  maxOutputBytes?: number;
};

export class EmberClientError extends Error {
  readonly code: "INVALID_INPUT" | "PROCESS_FAILED" | "TIMEOUT" | "INVALID_RESPONSE";
  readonly exitCode?: number;
  readonly stderr?: string;

  constructor(
    code: EmberClientError["code"],
    message: string,
    details: { exitCode?: number; stderr?: string; cause?: unknown } = {},
  ) {
    super(message, { cause: details.cause });
    this.name = "EmberClientError";
    this.code = code;
    this.exitCode = details.exitCode;
    this.stderr = details.stderr;
  }
}

export class EmberClient {
  readonly binaryPath: string;
  readonly cwd: string;
  readonly timeoutMs: number;
  readonly maxOutputBytes: number;

  constructor(options: ClientOptions) {
    if (!options.binaryPath?.trim()) {
      throw new EmberClientError("INVALID_INPUT", "binaryPath is required");
    }
    this.cwd = resolve(options.cwd ?? process.cwd());
    this.binaryPath = isAbsolute(options.binaryPath)
      ? options.binaryPath
      : resolve(this.cwd, options.binaryPath);
    this.timeoutMs = options.timeoutMs ?? 10_000;
    this.maxOutputBytes = options.maxOutputBytes ?? 4 * 1024 * 1024;
    if (!Number.isSafeInteger(this.timeoutMs) || this.timeoutMs < 100 || this.timeoutMs > 120_000) {
      throw new EmberClientError("INVALID_INPUT", "timeoutMs must be between 100 and 120000");
    }
    if (!Number.isSafeInteger(this.maxOutputBytes) || this.maxOutputBytes < 1_024) {
      throw new EmberClientError("INVALID_INPUT", "maxOutputBytes must be at least 1024");
    }
  }

  runScenario(scenarioPath: string, options: { includeEvents?: boolean } = {}): EmberReport {
    const path = this.scenarioPath(scenarioPath);
    const args = ["run", path, "--json"];
    if (options.includeEvents) args.push("--events");
    const output = this.execute(args);
    try {
      const parsed = JSON.parse(output) as unknown;
      assertReport(parsed);
      return parsed;
    } catch (error) {
      if (error instanceof EmberClientError) throw error;
      throw new EmberClientError("INVALID_RESPONSE", "EmberDTL returned malformed JSON", {
        cause: error,
      });
    }
  }

  validateScenario(scenarioPath: string): string[] {
    const output = this.execute(["validate", this.scenarioPath(scenarioPath), "--json"]);
    try {
      const parsed = JSON.parse(output) as { status?: unknown; messages?: unknown };
      if (parsed.status !== "ok" || !Array.isArray(parsed.messages)) {
        throw new Error("unexpected validation envelope");
      }
      return parsed.messages.map((message) => String(message));
    } catch (error) {
      throw new EmberClientError(
        "INVALID_RESPONSE",
        "EmberDTL returned an invalid validation response",
        {
          cause: error,
        },
      );
    }
  }

  private scenarioPath(value: string): string {
    if (!value?.trim()) {
      throw new EmberClientError("INVALID_INPUT", "scenario path is required");
    }
    const path = isAbsolute(value) ? value : resolve(this.cwd, value);
    if (!path.toLowerCase().endsWith(".json")) {
      throw new EmberClientError("INVALID_INPUT", "scenario must be a JSON file");
    }
    if (!existsSync(path) || !statSync(path).isFile()) {
      throw new EmberClientError("INVALID_INPUT", `scenario does not exist: ${path}`);
    }
    return path;
  }

  private execute(args: string[]): string {
    if (!existsSync(this.binaryPath) || !statSync(this.binaryPath).isFile()) {
      throw new EmberClientError("INVALID_INPUT", `binary does not exist: ${this.binaryPath}`);
    }
    const result = spawnSync(this.binaryPath, args, {
      cwd: this.cwd,
      encoding: "utf8",
      shell: false,
      timeout: this.timeoutMs,
      maxBuffer: this.maxOutputBytes,
      windowsHide: true,
    });
    if (result.error) {
      const timedOut = result.error.message.toLowerCase().includes("timeout");
      throw new EmberClientError(timedOut ? "TIMEOUT" : "PROCESS_FAILED", result.error.message, {
        exitCode: result.status ?? undefined,
        stderr: result.stderr,
        cause: result.error,
      });
    }
    if (result.status !== 0) {
      throw new EmberClientError("PROCESS_FAILED", "EmberDTL command failed", {
        exitCode: result.status ?? undefined,
        stderr: result.stderr.trim(),
      });
    }
    if (Buffer.byteLength(result.stdout, "utf8") > this.maxOutputBytes) {
      throw new EmberClientError(
        "INVALID_RESPONSE",
        "EmberDTL response exceeded the configured limit",
      );
    }
    return result.stdout;
  }
}

function assertReport(value: unknown): asserts value is EmberReport {
  if (typeof value !== "object" || value === null) {
    throw new EmberClientError("INVALID_RESPONSE", "report must be an object");
  }
  const report = value as Partial<EmberReport>;
  if (
    typeof report.name !== "string" ||
    !Number.isSafeInteger(report.epoch) ||
    !Array.isArray(report.assets) ||
    !Array.isArray(report.accounts) ||
    !Array.isArray(report.pools) ||
    !Array.isArray(report.reconciliation) ||
    typeof report.metrics !== "object" ||
    report.metrics === null
  ) {
    throw new EmberClientError("INVALID_RESPONSE", "report does not match the EmberDTL schema");
  }
}

export function amountOf(entries: AmountEntry[], asset: string): number {
  return entries.find((entry) => entry.id === asset)?.amount ?? 0;
}
