const api = (() => {
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
      const err = new Error(text.charAt(0).toUpperCase() + text.slice(1));
      err.status = res.status;
      throw err;
    }
    return data;
  }
  return {
    get: (url) => request('GET', url),
    post: (url, body) => request('POST', url, body ?? {}),
  };
})();

const formatMoney = (cents) =>
  new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(cents / 100);

function showMessage(el, text, type = 'error') {
  el.textContent = text;
  el.className = `message ${text ? type : ''}`;
}

document.querySelectorAll('#year').forEach((el) => (el.textContent = new Date().getFullYear()));
