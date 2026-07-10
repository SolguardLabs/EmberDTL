import test from "node:test";
import assert from "node:assert/strict";
import { amount, byId, runFixture, validateFixture } from "../helpers/runner.ts";

test("reserve deposits, facility settlement and fee contributions reconcile", () => {
  const report = runFixture("reserve_cycle.json");
  const reserve = byId(report.reserves, "ember-usd");
  const pool = byId(report.pools, "pool-usd");
  const facility = byId(report.facilities, "fac-1");
  const borrower = byId(report.accounts, "borrower");
  const merchant = byId(report.accounts, "merchant");
  const operator = byId(report.accounts, "operator");

  assert.equal(report.epoch, 1);
  assert.equal(reserve.balance, 10500);
  assert.equal(reserve.held, 1500);
  assert.equal(reserve.settlementVolume, 2500);
  assert.equal(reserve.accruedFees, 30);
  assert.equal(reserve.insuranceContributions, 15);
  assert.equal(pool.balance, 1015);
  assert.equal(pool.contributions, 1000);
  assert.equal(pool.feeContributions, 15);
  assert.equal(facility.outstanding, 1500);
  assert.equal(facility.repaid, 2500);
  assert.equal(facility.feesPaid, 30);
  assert.equal(amount(borrower.feesPaid, "usd"), 30);
  assert.equal(amount(merchant.settledIn, "usd"), 4000);
  assert.equal(amount(operator.free, "usd"), 15);
  assert.equal(validateFixture("reserve_cycle.json"), "ok");
});
