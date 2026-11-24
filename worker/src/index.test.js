import assert from "node:assert/strict";
import test from "node:test";

import {
  hashPassword,
  parseAmount,
  validateCardDetails,
  validateSignup,
  verifyPassword,
} from "./index.js";
import worker from "./index.js";

test("deposit and withdrawal amounts enforce whole-dollar limits", () => {
  assert.equal(parseAmount(20), 2_000);
  assert.equal(parseAmount(1_000_000), 100_000_000);
  assert.throws(() => parseAmount(19), /minimum amount.*\$20\.00/);
  assert.throws(() => parseAmount(20.5), /whole number/);
  assert.throws(() => parseAmount(1_000_001), /per-transaction limit/);
});

test("payment amounts allow a $1 minimum but reject fractions", () => {
  assert.equal(parseAmount(1, true), 100);
  assert.throws(() => parseAmount(0.5, true), /whole number/);
  assert.throws(() => parseAmount(0, true), /greater than 0/);
});

test("signup validates card and IBAN after normalizing separators", () => {
  validateSignup("Alice", "alice@example.com", "password123");
  validateCardDetails("5342 0000-0000 0001", "12/99", "eg96-mieb-0000000000000001");
  assert.throws(
    () => validateCardDetails("1234", "12/99", "EG96MIEB0000000000000001"),
    /exactly 16 digits/,
  );
  assert.throws(
    () => validateCardDetails("5342000000000001", "13/99", "EG96MIEB0000000000000001"),
    /MM\/YY/,
  );
  assert.throws(() => validateSignup("A", "not-an-email", "short"), /name must be/);
});

test("password hashes verify only the matching password", async () => {
  const hash = await hashPassword("password123");
  assert.match(hash, /^pbkdf2-sha256\$\d+\$/);
  assert.equal(await verifyPassword("password123", hash), true);
  assert.equal(await verifyPassword("wrong-password", hash), false);
});

test("protected API routes reject requests without a session", async () => {
  const response = await worker.fetch(new Request("https://miebank.test/api/balance"), { DB: {} });
  assert.equal(response.status, 401);
  assert.deepEqual(await response.json(), { error: "not logged in" });
});

test("API rejects unknown JSON fields", async () => {
  const response = await worker.fetch(
    new Request("https://miebank.test/api/signup", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name: "Alice", unknown: true }),
    }),
    { DB: {} },
  );
  assert.equal(response.status, 400);
  assert.deepEqual(await response.json(), { error: "invalid request body" });
});
