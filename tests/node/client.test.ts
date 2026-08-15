import assert from "node:assert/strict";
import test from "node:test";
import { EmberClient, EmberClientError, amountOf } from "../../sdk/client.ts";
import { binary, ensureBuilt, root } from "../helpers/runner.ts";

test("typed client executes a scenario without a command shell", () => {
  ensureBuilt();
  const client = new EmberClient({ binaryPath: binary, cwd: root });
  const report = client.runScenario("tests/fixtures/reserve_cycle.json");
  assert.equal(report.name, "ember reserve cycle");
  assert.equal(report.assets[0].id, "usd");
  assert.ok(report.reconciliation.length > 0);
});

test("typed client can include the journal", () => {
  ensureBuilt();
  const client = new EmberClient({ binaryPath: binary, cwd: root });
  const report = client.runScenario("tests/fixtures/claim_payment.json", { includeEvents: true });
  assert.ok(Array.isArray((report as unknown as { events: unknown[] }).events));
});

test("typed client validates a scenario", () => {
  ensureBuilt();
  const client = new EmberClient({ binaryPath: binary, cwd: root });
  assert.deepEqual(client.validateScenario("tests/fixtures/multi_asset.json"), ["ok"]);
});

test("typed client rejects non-JSON inputs before spawning", () => {
  const client = new EmberClient({ binaryPath: binary, cwd: root });
  assert.throws(
    () => client.runScenario("README.md"),
    (error: unknown) => error instanceof EmberClientError && error.code === "INVALID_INPUT",
  );
});

test("typed client rejects a missing binary", () => {
  const client = new EmberClient({ binaryPath: "bin/missing", cwd: root });
  assert.throws(
    () => client.runScenario("tests/fixtures/reserve_cycle.json"),
    (error: unknown) => error instanceof EmberClientError && error.code === "INVALID_INPUT",
  );
});

test("amountOf returns zero when an asset is absent", () => {
  assert.equal(amountOf([{ id: "usd", amount: 25 }], "usd"), 25);
  assert.equal(amountOf([{ id: "usd", amount: 25 }], "eur"), 0);
});
