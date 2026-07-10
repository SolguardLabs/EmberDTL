import { spawnSync } from "node:child_process";
import { existsSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

export type AmountEntry = {
  id: string;
  amount: number;
};

export type AccountReport = {
  id: string;
  role: string;
  status: string;
  free: AmountEntry[];
  held: AmountEntry[];
  settledIn: AmountEntry[];
  withdrawn: AmountEntry[];
  feesPaid: AmountEntry[];
  contributed: AmountEntry[];
  claimsPaid: AmountEntry[];
  defaultsTaken: AmountEntry[];
};

export type PoolReport = {
  id: string;
  asset: string;
  status: string;
  balance: number;
  contributions: number;
  feeContributions: number;
  claimsPaid: number;
  recovered: number;
  reservedForDefaults: number;
  pendingDefaults: number;
  pendingClaims: number;
};

export type ReserveReport = {
  id: string;
  asset: string;
  owner: string;
  status: string;
  balance: number;
  held: number;
  grossDeposits: number;
  grossWithdrawals: number;
  settlementVolume: number;
  accruedFees: number;
  insuranceContributions: number;
};

export type FacilityReport = {
  id: string;
  reserveId: string;
  asset: string;
  borrower: string;
  beneficiary: string;
  operator: string;
  status: string;
  principal: number;
  outstanding: number;
  repaid: number;
  feesPaid: number;
  openedEpoch: number;
  dueEpoch: number;
  defaultId?: string;
};

export type DefaultReport = {
  id: string;
  facilityId: string;
  asset: string;
  reporter: string;
  borrower: string;
  beneficiary: string;
  status: string;
  amount: number;
  expectedRecovery: number;
  coverageCeiling: number;
  pendingExposure: number;
  claimedCoverage: number;
  recovered: number;
  reportedEpoch: number;
  acceptedEpoch?: number;
  resolvedEpoch?: number;
  claims: string[];
};

export type ClaimReport = {
  id: string;
  defaultId: string;
  facilityId: string;
  asset: string;
  claimant: string;
  payoutAccount: string;
  status: string;
  amount: number;
  coverage: number;
  capacityAtRegistration: number;
  registeredEpoch: number;
  executedEpoch?: number;
};

export type EmberReport = {
  name: string;
  epoch: number;
  accounts: AccountReport[];
  pools: PoolReport[];
  reserves: ReserveReport[];
  facilities: FacilityReport[];
  defaults: DefaultReport[];
  claims: ClaimReport[];
  reconciliation: Array<Record<string, number | string>>;
  metrics: Record<string, number>;
  validations?: string[];
};

export const root = resolve(dirname(fileURLToPath(import.meta.url)), "..", "..");
export const binary = join(root, "bin", process.platform === "win32" ? "emberdtl.exe" : "emberdtl");

export function ensureBuilt(): void {
  if (existsSync(binary)) {
    return;
  }
  const result = spawnSync(process.execPath, ["scripts/build.mjs"], {
    cwd: root,
    encoding: "utf8",
  });
  if (result.status !== 0) {
    throw new Error(result.stderr || result.stdout || "build failed");
  }
}

export function runCli(args: string[]): string {
  ensureBuilt();
  const result = spawnSync(binary, args, {
    cwd: root,
    encoding: "utf8",
  });
  if (result.status !== 0) {
    throw new Error(`command failed: ${binary} ${args.join(" ")}\n${result.stderr}`);
  }
  return result.stdout;
}

export function runFixture(name: string): EmberReport {
  return JSON.parse(runCli(["run", join("tests", "fixtures", name), "--json"])) as EmberReport;
}

export function validateFixture(name: string): string {
  return runCli(["validate", join("tests", "fixtures", name)]).trim();
}

export function byId<T extends { id: string }>(items: T[], id: string): T {
  const found = items.find((item) => item.id === id);
  if (!found) {
    throw new Error(`missing ${id}`);
  }
  return found;
}

export function amount(entries: AmountEntry[], id: string): number {
  return entries.find((entry) => entry.id === id)?.amount ?? 0;
}
