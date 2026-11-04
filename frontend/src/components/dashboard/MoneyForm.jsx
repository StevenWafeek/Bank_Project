import { useState } from 'react';
import { BLOCKED_KEYS, QUICK_AMOUNTS, validateAmount } from '../../utils/amount.js';
import { formatMoney } from '../../utils/format.js';

const variants = {
  deposit: { title: 'Deposit', balanceLabel: 'Current balance', button: 'btn-success' },
  withdraw: { title: 'Withdraw', balanceLabel: 'Available', button: 'btn-danger' },
};

export default function MoneyForm({ type, balanceCents, onSubmit }) {
  const [amount, setAmount] = useState('');
  const [busy, setBusy] = useState(false);
  const v = variants[type];

  async function handleSubmit(e) {
    e.preventDefault();
    const error = validateAmount(amount);
    setBusy(true);
    const ok = await onSubmit(type, Number(amount), error);
    setBusy(false);
    if (ok) setAmount('');
  }

  return (
    <section className="card panel" id={`panel-${type}`}>
      <h2>{v.title}</h2>
      <p className="muted">
        {v.balanceLabel}: <strong className="balance-inline">{formatMoney(balanceCents)}</strong>
      </p>
      <form id={`${type}-form`} className="form inline-form" noValidate onSubmit={handleSubmit}>
        <label>
          Amount (USD) <small>(whole dollars, min. $20)</small>
          <input
            type="number"
            name="amount"
            min="20"
            step="1"
            inputMode="numeric"
            placeholder="20"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            onKeyDown={(e) => BLOCKED_KEYS.includes(e.key) && e.preventDefault()}
            required
          />
        </label>
        <div className="quick-amounts">
          {QUICK_AMOUNTS.map((q) => (
            <button key={q} type="button" data-amount={q} onClick={() => setAmount(String(q))}>
              ${q}
            </button>
          ))}
        </div>
        <button type="submit" className={`btn ${v.button}`} disabled={busy}>
          {v.title}
        </button>
      </form>
    </section>
  );
}
