import test from "node:test";
import assert from "node:assert/strict";
import { amount, byId, runFixture, validateFixture } from "../helpers/runner.ts";

test("insurance accounting remains separated across supported assets", () => {
  const report = runFixture("multi_asset.json");
  const usdPool = byId(report.pools, "pool-usd");
  const eurPool = byId(report.pools, "pool-eur");
  const usdReserve = byId(report.reserves, "usd-res");
  const eurReserve = byId(report.reserves, "eur-res");
  const usdFacility = byId(report.facilities, "fac-usd");
  const eurFacility = byId(report.facilities, "fac-eur");
  const borrower = byId(report.accounts, "borrower");
  const operator = byId(report.accounts, "operator");

  assert.equal(usdPool.balance, 1006);
  assert.equal(usdPool.feeContributions, 6);
  assert.equal(eurPool.balance, 500);
  assert.equal(eurPool.feeContributions, 0);
  assert.equal(usdReserve.balance, 6000);
  assert.equal(usdReserve.held, 2000);
  assert.equal(eurReserve.balance, 2000);
  assert.equal(eurReserve.held, 2000);
  assert.equal(usdFacility.outstanding, 2000);
  assert.equal(eurFacility.outstanding, 2000);
  assert.equal(amount(borrower.free, "usd"), 1988);
  assert.equal(amount(borrower.free, "eur"), 3000);
  assert.equal(amount(operator.free, "usd"), 6);
  assert.equal(amount(operator.free, "eur"), 0);
  assert.equal(validateFixture("multi_asset.json"), "ok");
});
