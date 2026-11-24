export class ApiError extends Error {
  constructor(message, status) {
    super(message);
    this.status = status;
  }
}

async function request(method, url, body) {
  const opts = { method, headers: {}, credentials: 'same-origin' };
  if (body !== undefined) {
    opts.headers['Content-Type'] = 'application/json';
    opts.body = JSON.stringify(body);
  }
  const res = await fetch(url, opts);
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    const text = data.error || `Request failed (${res.status})`;
    throw new ApiError(text.charAt(0).toUpperCase() + text.slice(1), res.status);
  }
  return data;
}

export const api = {
  me: () => request('GET', '/api/me'),
  signup: (name, email, password, card_number, card_expiry, iban) =>
    request('POST', '/api/signup', { name, email, password, card_number, card_expiry, iban }),
  login: (email, password) => request('POST', '/api/login', { email, password }),
  logout: () => request('POST', '/api/logout', {}),
  balance: () => request('GET', '/api/balance'),
  transactions: () => request('GET', '/api/transactions'),
  deposit: (amount) => request('POST', '/api/deposit', { amount }),
  withdraw: (amount) => request('POST', '/api/withdraw', { amount }),

  // New features
  transfer: (recipient, amount, note) => request('POST', '/api/transfer', { recipient, amount, note }),
  lookupRecipient: (query) => request('POST', '/api/transfer/lookup', { query }),
  payBill: (biller, reference, amount) => request('POST', '/api/bills/pay', { biller, reference, amount }),
  donate: (charity, amount, note, anonymous) => request('POST', '/api/donate', { charity, amount, note, anonymous }),
  updateProfile: (name) => request('PUT', '/api/account/profile', { name }),
  updateCard: (card_number, card_expiry, iban) =>
    request('PUT', '/api/account/card', { card_number, card_expiry, iban }),
  updatePassword: (current_password, new_password) =>
    request('PUT', '/api/account/password', { current_password, new_password }),
};
