import test from "node:test";
import assert from "node:assert/strict";
import { amount, byId, runFixture, validateFixture } from "../helpers/runner.ts";

test("default resolution records recovery and clears pending exposure", () => {
  const report = runFixture("default_resolution.json");
  const pool = byId(report.pools, "pool-usd");
  const def = byId(report.defaults, "def-recovery");
  const claim = byId(report.claims, "claim-recovery");
  const borrower = byId(report.accounts, "borrower");
  const merchant = byId(report.accounts, "merchant");

  assert.equal(report.epoch, 2);
  assert.equal(def.status, "resolved");
  assert.equal(def.pendingExposure, 0);
  assert.equal(def.recovered, 100);
  assert.equal(claim.status, "executed");
  assert.equal(claim.coverage, 400);
  assert.equal(pool.balance, 1700);
  assert.equal(pool.claimsPaid, 400);
  assert.equal(pool.recovered, 100);
  assert.equal(pool.pendingDefaults, 0);
  assert.equal(amount(borrower.free, "usd"), 2900);
  assert.equal(amount(merchant.claimsPaid, "usd"), 400);
  assert.equal(validateFixture("default_resolution.json"), "ok");
});
