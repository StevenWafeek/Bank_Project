import QRCode from "qrcode";

const COOKIE_NAME = "miebank_session";
const SESSION_TTL_SECONDS = 30 * 60;
const PASSWORD_ITERATIONS = 210_000;
const MAX_BODY_BYTES = 64 * 1024;
const MAX_AMOUNT_CENTS = 100_000_000;
const USER_COLUMNS =
  "id, name, email, password_hash, account_number, card_number, card_expiry, iban, balance_cents, created_at";

class HttpError extends Error {
  constructor(status, message) {
    super(message);
    this.status = status;
  }
}

function jsonResponse(value, status = 200, headers = {}) {
  return new Response(JSON.stringify(value), {
    status,
    headers: {
      "Content-Type": "application/json; charset=utf-8",
      "Cache-Control": "no-store",
      ...headers,
    },
  });
}

function errorResponse(status, message) {
  return jsonResponse({ error: message }, status);
}

async function readJson(request, allowedKeys) {
  const text = await request.text();
  if (new TextEncoder().encode(text).byteLength > MAX_BODY_BYTES) {
    throw new HttpError(400, "invalid request body");
  }

  let value;
  try {
    value = JSON.parse(text);
  } catch {
    throw new HttpError(400, "invalid request body");
  }
  if (!value || Array.isArray(value) || typeof value !== "object") {
    throw new HttpError(400, "invalid request body");
  }
  if (Object.keys(value).some((key) => !allowedKeys.includes(key))) {
    throw new HttpError(400, "invalid request body");
  }
  return value;
}

function stringValue(value) {
  return typeof value === "string" ? value : "";
}

function textLength(value) {
  return new TextEncoder().encode(value).length;
}

function cleanCardNumber(value) {
  return value.trim().replace(/[^0-9]/g, "");
}

function cleanIBAN(value) {
  return value.trim().toUpperCase().replace(/[^A-Z0-9]/g, "");
}

function validateSignup(name, email, password) {
  if (textLength(name) < 2 || textLength(name) > 60) {
    throw new HttpError(400, "name must be between 2 and 60 characters");
  }
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
    throw new HttpError(400, "please enter a valid email address");
  }
  if (textLength(password) < 8 || textLength(password) > 72) {
    throw new HttpError(400, "password must be between 8 and 72 characters");
  }
}

function validateCardDetails(cardNumber, cardExpiry, iban) {
  if (!/^\d{16}$/.test(cleanCardNumber(cardNumber))) {
    throw new HttpError(400, "card number must be exactly 16 digits");
  }
  if (!/^(0[1-9]|1[0-2])\/\d{2}$/.test(cardExpiry)) {
    throw new HttpError(400, "card expiry must be in MM/YY format (e.g. 08/28)");
  }

  const [month, shortYear] = cardExpiry.split("/").map(Number);
  if (Date.UTC(2000 + shortYear, month, 1) <= Date.now()) {
    throw new HttpError(400, "card expiry date has already passed");
  }
  if (!/^[A-Z]{2}\d{2}[A-Z0-9]{1,30}$/.test(cleanIBAN(iban))) {
    throw new HttpError(
      400,
      "IBAN must start with 2 letters and 2 digits followed by up to 30 alphanumeric characters (e.g. EG96MIEB0000000012345678)",
    );
  }
}

function bytesToBase64(bytes) {
  let binary = "";
  for (const byte of bytes) binary += String.fromCharCode(byte);
  return btoa(binary);
}

function base64ToBytes(value) {
  return Uint8Array.from(atob(value), (char) => char.charCodeAt(0));
}

function randomHex(byteLength) {
  const bytes = crypto.getRandomValues(new Uint8Array(byteLength));
  return Array.from(bytes, (byte) => byte.toString(16).padStart(2, "0")).join("");
}

function randomDigits(length) {
  const bytes = crypto.getRandomValues(new Uint8Array(length));
  return Array.from(bytes, (byte) => String(byte % 10)).join("");
}

function randomInteger(maxExclusive) {
  const value = new Uint32Array(1);
  crypto.getRandomValues(value);
  return value[0] % maxExclusive;
}

async function hashPassword(password) {
  const salt = crypto.getRandomValues(new Uint8Array(16));
  const key = await crypto.subtle.importKey(
    "raw",
    new TextEncoder().encode(password),
    "PBKDF2",
    false,
    ["deriveBits"],
  );
  const hash = new Uint8Array(
    await crypto.subtle.deriveBits(
      { name: "PBKDF2", hash: "SHA-256", salt, iterations: PASSWORD_ITERATIONS },
      key,
      256,
    ),
  );
  return `pbkdf2-sha256$${PASSWORD_ITERATIONS}$${bytesToBase64(salt)}$${bytesToBase64(hash)}`;
}

async function verifyPassword(password, encodedHash) {
  const [algorithm, iterationsText, saltText, expectedText] = encodedHash.split("$");
  const iterations = Number(iterationsText);
  if (algorithm !== "pbkdf2-sha256" || !Number.isSafeInteger(iterations) || iterations < 1) {
    return false;
  }

  const key = await crypto.subtle.importKey(
    "raw",
    new TextEncoder().encode(password),
    "PBKDF2",
    false,
    ["deriveBits"],
  );
  const actual = new Uint8Array(
    await crypto.subtle.deriveBits(
      { name: "PBKDF2", hash: "SHA-256", salt: base64ToBytes(saltText), iterations },
      key,
      256,
    ),
  );
  const expected = base64ToBytes(expectedText);
  if (actual.length !== expected.length) return false;
  let difference = 0;
  for (let index = 0; index < actual.length; index += 1) {
    difference |= actual[index] ^ expected[index];
  }
  return difference === 0;
}

async function sha256(value) {
  const digest = await crypto.subtle.digest("SHA-256", new TextEncoder().encode(value));
  return bytesToBase64(new Uint8Array(digest));
}

function parseAmount(value, payment = false) {
  if (typeof value !== "number" || !Number.isFinite(value) || value <= 0) {
    throw new HttpError(400, "amount must be greater than 0");
  }
  if (!Number.isInteger(value)) {
    throw new HttpError(400, "amount must be a whole number (no decimals)");
  }
  const cents = value * 100;
  const minimum = payment ? 100 : 2_000;
  if (cents < minimum) {
    throw new HttpError(
      400,
      payment
        ? "the minimum amount is $1.00"
        : "the minimum amount for a deposit or withdrawal is $20.00",
    );
  }
  if (cents > MAX_AMOUNT_CENTS) {
    throw new HttpError(400, "amount exceeds the per-transaction limit of 1,000,000.00");
  }
  return cents;
}

function toUserResponse(user) {
  return {
    id: Number(user.id),
    name: user.name,
    email: user.email,
    account_number: user.account_number,
    card_number: user.card_number,
    card_expiry: user.card_expiry,
    iban: user.iban,
    balance_cents: Number(user.balance_cents),
    created_at: user.created_at,
  };
}

function toTransaction(row) {
  return {
    id: Number(row.id),
    type: row.type,
    amount_cents: Number(row.amount_cents),
    balance_cents: Number(row.balance_cents),
    description: row.description || "",
    created_at: row.created_at,
  };
}

function cookieHeader(token, request) {
  const secure = new URL(request.url).protocol === "https:" ? "; Secure" : "";
  return `${COOKIE_NAME}=${token}; Path=/; HttpOnly; SameSite=Strict; Max-Age=${SESSION_TTL_SECONDS}${secure}`;
}

function expiredCookie(request) {
  const secure = new URL(request.url).protocol === "https:" ? "; Secure" : "";
  return `${COOKIE_NAME}=; Path=/; HttpOnly; SameSite=Strict; Max-Age=0${secure}`;
}

function getCookie(request, name) {
  const cookies = request.headers.get("Cookie") || "";
  for (const part of cookies.split(";")) {
    const separator = part.indexOf("=");
    if (separator < 0) continue;
    if (part.slice(0, separator).trim() === name) {
      return part.slice(separator + 1).trim();
    }
  }
  return "";
}

async function getUserBy(db, condition, value) {
  return db.prepare(`SELECT ${USER_COLUMNS} FROM users WHERE ${condition} LIMIT 1`).bind(value).first();
}

async function getUserByIdentifier(db, value) {
  const trimmed = value.trim();
  return db
    .prepare(
      `SELECT ${USER_COLUMNS} FROM users
       WHERE email = ? COLLATE NOCASE
          OR account_number = ?
          OR (card_number != '' AND card_number = ?)
          OR (iban != '' AND iban = ?)
       LIMIT 1`,
    )
    .bind(trimmed.toLowerCase(), trimmed.toUpperCase(), cleanCardNumber(trimmed), cleanIBAN(trimmed))
    .first();
}

async function createSession(db, userId) {
  const token = randomHex(32);
  const tokenHash = await sha256(token);
  const expiry = new Date(Date.now() + SESSION_TTL_SECONDS * 1000).toISOString();
  await db
    .prepare("INSERT INTO sessions (token_hash, user_id, expires_at) VALUES (?, ?, ?)")
    .bind(tokenHash, userId, expiry)
    .run();
  return token;
}

async function requireUser(request, env) {
  const token = getCookie(request, COOKIE_NAME);
  if (!token) throw new HttpError(401, "not logged in");

  const tokenHash = await sha256(token);
  const now = new Date().toISOString();
  const session = await env.DB
    .prepare("SELECT user_id FROM sessions WHERE token_hash = ? AND expires_at > ?")
    .bind(tokenHash, now)
    .first();
  if (!session) {
    await env.DB.prepare("DELETE FROM sessions WHERE token_hash = ?").bind(tokenHash).run();
    throw new HttpError(401, "session expired, please log in again");
  }

  const expiry = new Date(Date.now() + SESSION_TTL_SECONDS * 1000).toISOString();
  await env.DB
    .prepare("UPDATE sessions SET expires_at = ? WHERE token_hash = ?")
    .bind(expiry, tokenHash)
    .run();

  const user = await getUserBy(env.DB, "id = ?", session.user_id);
  if (!user) throw new HttpError(401, "not logged in");
  return user;
}

async function createUser(request, env) {
  const body = await readJson(request, [
    "name",
    "email",
    "password",
    "card_number",
    "card_expiry",
    "iban",
  ]);
  const name = stringValue(body.name).trim();
  const email = stringValue(body.email).trim().toLowerCase();
  const password = stringValue(body.password);
  const cardNumber = cleanCardNumber(stringValue(body.card_number));
  const cardExpiry = stringValue(body.card_expiry).trim();
  const iban = cleanIBAN(stringValue(body.iban));

  validateSignup(name, email, password);
  validateCardDetails(cardNumber, cardExpiry, iban);

  const passwordHash = await hashPassword(password);
  const accountNumber = `MIE${randomDigits(10)}`;
  const createdAt = new Date().toISOString();
  let user;
  try {
    user = await env.DB
      .prepare(
        `INSERT INTO users
         (name, email, password_hash, account_number, card_number, card_expiry, iban, created_at)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?)
         RETURNING ${USER_COLUMNS}`,
      )
      .bind(name, email, passwordHash, accountNumber, cardNumber, cardExpiry, iban, createdAt)
      .first();
  } catch (error) {
    const detail = String(error);
    if (detail.includes("users.email")) throw new HttpError(409, "email already registered");
    if (detail.includes("users.card_number")) {
      throw new HttpError(409, "this card number is already associated with another account");
    }
    if (detail.includes("users.iban")) {
      throw new HttpError(409, "this IBAN is already associated with another account");
    }
    throw error;
  }

  const token = await createSession(env.DB, user.id);
  return jsonResponse(toUserResponse(user), 201, { "Set-Cookie": cookieHeader(token, request) });
}

async function login(request, env) {
  const body = await readJson(request, ["email", "identifier", "password"]);
  const identifier = stringValue(body.email || body.identifier).trim();
  const password = stringValue(body.password);
  if (!identifier) throw new HttpError(400, "please enter your email or card number");

  const user = await getUserByIdentifier(env.DB, identifier);
  if (!user || !(await verifyPassword(password, user.password_hash))) {
    throw new HttpError(401, "invalid email or password");
  }
  const token = await createSession(env.DB, user.id);
  return jsonResponse(toUserResponse(user), 200, { "Set-Cookie": cookieHeader(token, request) });
}

async function logout(request, env) {
  const token = getCookie(request, COOKIE_NAME);
  if (token) {
    await env.DB.prepare("DELETE FROM sessions WHERE token_hash = ?").bind(await sha256(token)).run();
  }
  return jsonResponse(
    { status: "logged out" },
    200,
    { "Set-Cookie": expiredCookie(request) },
  );
}

async function getTransaction(db, id, userId) {
  const row = await db
    .prepare(
      `SELECT id, type, amount_cents, balance_cents, description, created_at
       FROM transactions WHERE id = ? AND user_id = ?`,
    )
    .bind(id, userId)
    .first();
  if (!row) throw new Error("transaction was not recorded");
  return toTransaction(row);
}

async function applyTransaction(db, userId, type, delta, description = "") {
  const createdAt = new Date().toISOString();
  const results = await db.batch([
    db
      .prepare(
        `UPDATE users SET balance_cents = balance_cents + ?
         WHERE id = ? AND balance_cents + ? >= 0`,
      )
      .bind(delta, userId, delta),
    db
      .prepare(
        `INSERT INTO transactions (user_id, type, amount_cents, balance_cents, description, created_at)
         SELECT ?, ?, ?, balance_cents, ?, ? FROM users WHERE id = ? AND changes() > 0`,
      )
      .bind(userId, type, Math.abs(delta), description, createdAt, userId),
  ]);

  if (results[0].meta.changes === 0) {
    if (delta < 0) throw new HttpError(422, "insufficient funds");
    throw new Error("account balance update failed");
  }
  const id = Number(results[1].meta.last_row_id);
  const transaction = await getTransaction(db, id, userId);
  return transaction;
}

async function moneyTransaction(request, env, user, type, direction) {
  const body = await readJson(request, ["amount"]);
  const cents = parseAmount(body.amount);
  const transaction = await applyTransaction(env.DB, user.id, type, direction * cents);
  return jsonResponse({
    transaction,
    balance_cents: transaction.balance_cents,
  });
}

async function listTransactions(env, user) {
  const result = await env.DB
    .prepare(
      `SELECT id, type, amount_cents, balance_cents, description, created_at
       FROM transactions WHERE user_id = ? ORDER BY id DESC LIMIT 50`,
    )
    .bind(user.id)
    .all();
  return jsonResponse({ transactions: result.results.map(toTransaction) });
}

async function findRecipient(db, query) {
  const trimmed = query.trim();
  return db
    .prepare(
      `SELECT id, name, account_number, email, card_number, iban, balance_cents
       FROM users
       WHERE account_number = ?
          OR email = ? COLLATE NOCASE
          OR (card_number != '' AND card_number = ?)
          OR (iban != '' AND iban = ?)
       LIMIT 1`,
    )
    .bind(trimmed.toUpperCase(), trimmed.toLowerCase(), cleanCardNumber(trimmed), cleanIBAN(trimmed))
    .first();
}

async function transfer(request, env, user) {
  const body = await readJson(request, ["recipient", "amount", "note"]);
  const query = stringValue(body.recipient).trim();
  if (!query) throw new HttpError(400, "recipient account number or email is required");
  const cents = parseAmount(body.amount, true);
  const recipient = await findRecipient(env.DB, query);
  if (!recipient) throw new HttpError(404, "recipient account not found");
  if (Number(recipient.id) === Number(user.id)) {
    throw new HttpError(400, "cannot transfer to your own account");
  }

  const note = stringValue(body.note).trim();
  const memo = note ? ` - "${note}"` : "";
  const createdAt = new Date().toISOString();
  const results = await env.DB.batch([
    env.DB
      .prepare("UPDATE users SET balance_cents = balance_cents - ? WHERE id = ? AND balance_cents >= ?")
      .bind(cents, user.id, cents),
    env.DB
      .prepare("UPDATE users SET balance_cents = balance_cents + ? WHERE id = ? AND changes() > 0")
      .bind(cents, recipient.id),
    env.DB
      .prepare(
        `INSERT INTO transactions (user_id, type, amount_cents, balance_cents, description, created_at)
         SELECT ?, 'transfer_out', ?, balance_cents, ?, ? FROM users WHERE id = ? AND changes() > 0`,
      )
      .bind(
        user.id,
        cents,
        `Transfer to ${recipient.name} (${recipient.account_number})${memo}`,
        createdAt,
        user.id,
      ),
    env.DB
      .prepare(
        `INSERT INTO transactions (user_id, type, amount_cents, balance_cents, description, created_at)
         SELECT ?, 'transfer_in', ?, balance_cents, ?, ? FROM users WHERE id = ? AND changes() > 0`,
      )
      .bind(
        recipient.id,
        cents,
        `Transfer from ${user.name} (${user.account_number})${memo}`,
        createdAt,
        recipient.id,
      ),
  ]);

  if (results[0].meta.changes === 0) {
    throw new HttpError(422, "insufficient funds for transfer");
  }
  const transaction = await getTransaction(
    env.DB,
    Number(results[2].meta.last_row_id),
    user.id,
  );
  return jsonResponse({
    transaction,
    balance_cents: transaction.balance_cents,
    recipient_name: recipient.name,
    recipient_account: recipient.account_number,
  });
}

async function transferLookup(request, env, user) {
  const body = await readJson(request, ["query"]);
  const query = stringValue(body.query).trim();
  if (!query) throw new HttpError(400, "please enter an account number or email");
  const recipient = await findRecipient(env.DB, query);
  if (!recipient) throw new HttpError(404, "recipient not found");
  if (Number(recipient.id) === Number(user.id)) {
    throw new HttpError(400, "cannot transfer to your own account");
  }
  return jsonResponse({ name: recipient.name, account_number: recipient.account_number });
}

async function payment(request, env, user, type) {
  if (type === "bill_payment") {
    const body = await readJson(request, ["biller", "reference", "amount"]);
    const biller = stringValue(body.biller).trim();
    const reference = stringValue(body.reference).trim();
    if (biller.length < 2 || biller.length > 80) {
      throw new HttpError(400, "please select or enter a valid biller name");
    }
    if (reference.length < 2 || reference.length > 60) {
      throw new HttpError(400, "please enter a valid account or reference number");
    }
    const cents = parseAmount(body.amount, true);
    const transaction = await applyTransaction(
      env.DB,
      user.id,
      type,
      -cents,
      `Bill Payment: ${biller} (Ref: ${reference})`,
    );
    return jsonResponse({
      transaction,
      balance_cents: transaction.balance_cents,
      confirmation_code: `BILL-${10_000_000 + randomInteger(90_000_000)}`,
      biller,
      reference,
    });
  }

  const body = await readJson(request, ["charity", "amount", "note", "anonymous"]);
  const charity = stringValue(body.charity).trim();
  if (charity.length < 2 || charity.length > 80) {
    throw new HttpError(400, "please select or enter a valid charity name");
  }
  const cents = parseAmount(body.amount, true);
  const note = stringValue(body.note).trim();
  const description = `Donation to ${charity}${note ? ` - "${note}"` : ""}`;
  const transaction = await applyTransaction(env.DB, user.id, type, -cents, description);
  return jsonResponse({
    transaction,
    balance_cents: transaction.balance_cents,
    receipt_id: `DON-${10_000_000 + randomInteger(90_000_000)}`,
    charity,
  });
}

async function updateProfile(request, env, user) {
  const body = await readJson(request, ["name"]);
  const name = stringValue(body.name).trim();
  if (textLength(name) < 2 || textLength(name) > 60) {
    throw new HttpError(400, "name must be between 2 and 60 characters");
  }
  await env.DB.prepare("UPDATE users SET name = ? WHERE id = ?").bind(name, user.id).run();
  return jsonResponse(toUserResponse({ ...user, name }));
}

async function updatePassword(request, env, user) {
  const body = await readJson(request, ["current_password", "new_password"]);
  const currentPassword = stringValue(body.current_password);
  const newPassword = stringValue(body.new_password);
  if (!(await verifyPassword(currentPassword, user.password_hash))) {
    throw new HttpError(401, "incorrect current password");
  }
  if (textLength(newPassword) < 8 || textLength(newPassword) > 72) {
    throw new HttpError(400, "new password must be between 8 and 72 characters");
  }
  const passwordHash = await hashPassword(newPassword);
  await env.DB
    .prepare("UPDATE users SET password_hash = ? WHERE id = ?")
    .bind(passwordHash, user.id)
    .run();
  return jsonResponse({ message: "password updated successfully" });
}

async function updateCard(request, env, user) {
  const body = await readJson(request, ["card_number", "card_expiry", "iban"]);
  const cardNumber = cleanCardNumber(stringValue(body.card_number));
  const cardExpiry = stringValue(body.card_expiry).trim();
  const iban = cleanIBAN(stringValue(body.iban));
  validateCardDetails(cardNumber, cardExpiry, iban);

  try {
    await env.DB
      .prepare("UPDATE users SET card_number = ?, card_expiry = ?, iban = ? WHERE id = ?")
      .bind(cardNumber, cardExpiry, iban, user.id)
      .run();
  } catch (error) {
    const detail = String(error);
    if (detail.includes("users.card_number")) {
      throw new HttpError(409, "this card number is already associated with another account");
    }
    if (detail.includes("users.iban")) {
      throw new HttpError(409, "this IBAN is already associated with another account");
    }
    throw error;
  }

  return jsonResponse(
    toUserResponse({ ...user, card_number: cardNumber, card_expiry: cardExpiry, iban }),
  );
}

async function accountQR(env, user) {
  const payload = JSON.stringify({
    bank: "MIeBank",
    account_number: user.account_number,
    card_number: user.card_number,
    iban: user.iban,
    name: user.name,
  });
  const png = await QRCode.toBuffer(payload, { errorCorrectionLevel: "M", width: 256 });
  return new Response(png, {
    status: 200,
    headers: {
      "Content-Type": "image/png",
      "Cache-Control": "private, max-age=60",
    },
  });
}

const protectedRoutes = new Set([
  "GET /api/me",
  "GET /api/balance",
  "POST /api/deposit",
  "POST /api/withdraw",
  "GET /api/transactions",
  "POST /api/transfer",
  "POST /api/transfer/lookup",
  "POST /api/bills/pay",
  "POST /api/donate",
  "PUT /api/account/profile",
  "PUT /api/account/password",
  "PUT /api/account/card",
  "GET /api/account/qr",
]);

const knownRoutes = new Set([
  ...protectedRoutes,
  "POST /api/signup",
  "POST /api/login",
  "POST /api/logout",
]);

async function handleApi(request, env) {
  const { pathname } = new URL(request.url);
  const route = `${request.method} ${pathname}`;
  if (!knownRoutes.has(route)) {
    const pathExists = [...knownRoutes].some((item) => item.endsWith(` ${pathname}`));
    return pathExists
      ? errorResponse(405, "method not allowed")
      : errorResponse(404, "not found");
  }

  try {
    if (route === "POST /api/signup") return await createUser(request, env);
    if (route === "POST /api/login") return await login(request, env);
    if (route === "POST /api/logout") return await logout(request, env);

    const user = await requireUser(request, env);
    if (route === "GET /api/me") return jsonResponse(toUserResponse(user));
    if (route === "GET /api/balance") {
      return jsonResponse({
        balance_cents: Number(user.balance_cents),
        account_number: user.account_number,
        card_number: user.card_number,
        card_expiry: user.card_expiry,
        iban: user.iban,
      });
    }
    if (route === "POST /api/deposit") {
      return await moneyTransaction(request, env, user, "deposit", 1);
    }
    if (route === "POST /api/withdraw") {
      return await moneyTransaction(request, env, user, "withdraw", -1);
    }
    if (route === "GET /api/transactions") return await listTransactions(env, user);
    if (route === "POST /api/transfer") return await transfer(request, env, user);
    if (route === "POST /api/transfer/lookup") {
      return await transferLookup(request, env, user);
    }
    if (route === "POST /api/bills/pay") {
      return await payment(request, env, user, "bill_payment");
    }
    if (route === "POST /api/donate") {
      return await payment(request, env, user, "donation");
    }
    if (route === "PUT /api/account/profile") return await updateProfile(request, env, user);
    if (route === "PUT /api/account/password") return await updatePassword(request, env, user);
    if (route === "PUT /api/account/card") return await updateCard(request, env, user);
    if (route === "GET /api/account/qr") return await accountQR(env, user);
  } catch (error) {
    if (error instanceof HttpError) return errorResponse(error.status, error.message);
    console.error("MIeBank API request failed", error);
    return errorResponse(500, "internal error");
  }
  return errorResponse(404, "not found");
}

export default {
  async fetch(request, env) {
    if (new URL(request.url).pathname.startsWith("/api/")) {
      return handleApi(request, env);
    }
    return env.ASSETS.fetch(request);
  },
};

export { hashPassword, parseAmount, validateCardDetails, validateSignup, verifyPassword };
