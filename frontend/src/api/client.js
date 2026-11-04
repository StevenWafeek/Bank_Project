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
  signup: (name, email, password) => request('POST', '/api/signup', { name, email, password }),
  login: (email, password) => request('POST', '/api/login', { email, password }),
  logout: () => request('POST', '/api/logout', {}),
  balance: () => request('GET', '/api/balance'),
  transactions: () => request('GET', '/api/transactions'),
  deposit: (amount) => request('POST', '/api/deposit', { amount }),
  withdraw: (amount) => request('POST', '/api/withdraw', { amount }),
};
