import test from "node:test";
import assert from "node:assert/strict";
import { amount, byId, runFixture, validateFixture } from "../helpers/runner.ts";

test("default claim payment updates pool, claim and beneficiary reports", () => {
  const report = runFixture("claim_payment.json");
  const pool = byId(report.pools, "pool-usd");
  const facility = byId(report.facilities, "fac-loss");
  const def = byId(report.defaults, "def-1");
  const claim = byId(report.claims, "claim-1");
  const merchant = byId(report.accounts, "merchant");

  assert.equal(report.epoch, 2);
  assert.equal(def.status, "accepted");
  assert.equal(def.amount, 2000);
  assert.equal(def.coverageCeiling, 1600);
  assert.equal(def.pendingExposure, 200);
  assert.deepEqual(def.claims, ["claim-1"]);
  assert.equal(claim.status, "executed");
  assert.equal(claim.amount, 1000);
  assert.equal(claim.coverage, 800);
  assert.equal(claim.capacityAtRegistration, 2000);
  assert.equal(pool.balance, 2200);
  assert.equal(pool.claimsPaid, 800);
  assert.equal(pool.pendingDefaults, 200);
  assert.equal(pool.pendingClaims, 0);
  assert.equal(facility.status, "defaulted");
  assert.equal(facility.outstanding, 4200);
  assert.equal(amount(merchant.claimsPaid, "usd"), 800);
  assert.equal(amount(merchant.free, "usd"), 5800);
  assert.equal(validateFixture("claim_payment.json"), "ok");
});
