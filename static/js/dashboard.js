(() => {
  const MIN_AMOUNT = 20;
  const msg = document.getElementById('dash-message');

  function renderBalance(cents) {
    document.getElementById('balance-amount').textContent = formatMoney(cents);
    document.querySelectorAll('.balance-inline').forEach((el) => (el.textContent = formatMoney(cents)));
  }

  function renderTransactions(txs) {
    const body = document.getElementById('tx-body');
    body.innerHTML = '';
    if (!txs.length) {
      body.innerHTML = '<tr><td colspan="4" class="muted center">No transactions yet.</td></tr>';
      return;
    }
    for (const tx of txs) {
      const tr = document.createElement('tr');
      const sign = tx.type === 'deposit' ? '+' : '-';
      const cells = [
        new Date(tx.created_at).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' }),
        tx.type === 'deposit' ? 'Deposit' : 'Withdraw',
        `${sign}${formatMoney(tx.amount_cents)}`,
        formatMoney(tx.balance_cents),
      ];
      cells.forEach((text, i) => {
        const td = document.createElement('td');
        td.textContent = text;
        if (i >= 2) td.className = 'num';
        if (i === 2) td.classList.add(tx.type === 'deposit' ? 'pos' : 'neg');
        tr.appendChild(td);
      });
      body.appendChild(tr);
    }
  }

  async function refresh() {
    const [bal, txs] = await Promise.all([api.get('/api/balance'), api.get('/api/transactions')]);
    renderBalance(bal.balance_cents);
    renderTransactions(txs.transactions);
  }

  function handleAuthError(err) {
    if (err.status === 401) {
      location.replace('/');
      return true;
    }
    return false;
  }

  async function init() {
    try {
      const me = await api.get('/api/me');
      document.getElementById('user-name').textContent = me.name;
      document.getElementById('account-number').textContent = me.account_number;
      await refresh();
    } catch (err) {
      if (!handleAuthError(err)) showMessage(msg, err.message);
    }
  }

  document.querySelectorAll('.option').forEach((opt) => {
    opt.addEventListener('click', () => {
      document.querySelectorAll('.option').forEach((o) => o.classList.toggle('active', o === opt));
      document.querySelectorAll('.panel').forEach((p) => p.classList.add('hidden'));
      document.getElementById(`panel-${opt.dataset.panel}`).classList.remove('hidden');
      showMessage(msg, '');
      if (opt.dataset.panel === 'balance') refresh().catch(handleAuthError);
    });
  });

  document.querySelectorAll('input[name=amount]').forEach((input) => {
    input.addEventListener('keydown', (e) => {
      if (['.', ',', 'e', 'E', '+', '-'].includes(e.key)) e.preventDefault();
    });
  });

  document.querySelectorAll('.quick-amounts button').forEach((btn) => {
    btn.addEventListener('click', () => {
      btn.closest('form').querySelector('input[name=amount]').value = btn.dataset.amount;
    });
  });

  function bindMoneyForm(formId, endpoint, verb) {
    const form = document.getElementById(formId);
    form.addEventListener('submit', async (e) => {
      e.preventDefault();
      const input = form.querySelector('input[name=amount]');
      const amount = Number(input.value);
      if (!input.value || !Number.isFinite(amount) || amount <= 0) {
        showMessage(msg, 'Please enter a valid amount greater than 0.');
        return;
      }
      if (!Number.isInteger(amount)) {
        showMessage(msg, 'Please enter a whole number (no decimals).');
        return;
      }
      if (amount < MIN_AMOUNT) {
        showMessage(msg, `The minimum amount for a deposit or withdrawal is ${formatMoney(MIN_AMOUNT * 100)}.`);
        return;
      }
      const btn = form.querySelector('button[type=submit]');
      btn.disabled = true;
      try {
        const res = await api.post(endpoint, { amount });
        input.value = '';
        showMessage(msg, `${verb} ${formatMoney(res.transaction.amount_cents)} successfully. New balance: ${formatMoney(res.balance_cents)}`, 'success');
        await refresh();
      } catch (err) {
        if (!handleAuthError(err)) showMessage(msg, err.message);
      } finally {
        btn.disabled = false;
      }
    });
  }

  bindMoneyForm('deposit-form', '/api/deposit', 'Deposited');
  bindMoneyForm('withdraw-form', '/api/withdraw', 'Withdrew');

  document.getElementById('logout-btn').addEventListener('click', async () => {
    await api.post('/api/logout').catch(() => {});
    location.replace('/');
  });

  init();
})();
