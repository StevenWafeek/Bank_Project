import { useState } from 'react';
import { BLOCKED_KEYS, validatePaymentAmount } from '../../utils/amount.js';
import { formatMoney } from '../../utils/format.js';

const POPULAR_BILLERS = [
  { id: 'electric', name: 'National Grid / City Electric', category: 'Electric & Gas' },
  { id: 'water', name: 'Municipal Water & Sewer', category: 'Water & Utilities' },
  { id: 'internet', name: 'Gigabit Broadband & Fiber', category: 'Internet & TV' },
  { id: 'telecom', name: 'Metro Wireless & Telecom', category: 'Mobile Phone' },
  { id: 'health', name: 'Horizon Health Insurance', category: 'Healthcare' },
];

const QUICK_AMOUNTS = [30, 50, 80, 120, 200];

export default function BillsPanel({ balanceCents, onSubmit }) {
  const [selectedBiller, setSelectedBiller] = useState(POPULAR_BILLERS[0].name);
  const [customBiller, setCustomBiller] = useState('');
  const [reference, setReference] = useState('');
  const [amount, setAmount] = useState('');
  const [loading, setLoading] = useState(false);

  const billerName = selectedBiller === 'custom' ? customBiller : selectedBiller;

  async function handleSubmit(e) {
    e.preventDefault();
    if (!billerName.trim()) {
      onSubmit('Please specify a biller name.');
      return;
    }
    if (!reference.trim()) {
      onSubmit('Please enter your account or bill reference number.');
      return;
    }
    const err = validatePaymentAmount(amount, 1);
    if (err) {
      onSubmit(err);
      return;
    }
    const cents = Number(amount) * 100;
    if (cents > balanceCents) {
      onSubmit('Insufficient funds to pay this bill.');
      return;
    }

    setLoading(true);
    const ok = await onSubmit(null, {
      biller: billerName.trim(),
      reference: reference.trim(),
      amount: Number(amount),
    });
    setLoading(false);
    if (ok) {
      setAmount('');
      setReference('');
      if (selectedBiller === 'custom') setCustomBiller('');
    }
  }

  return (
    <section className="card panel" id="panel-bills">
      <h2>Bill Payment</h2>
      <p className="muted">
        Pay utilities, internet, mobile and insurance bills directly. Available balance:{' '}
        <strong className="accent">{formatMoney(balanceCents)}</strong>
      </p>

      <form className="form inline-form" onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="biller-select">Select Biller</label>
          <select
            id="biller-select"
            className="form-select"
            value={selectedBiller}
            onChange={(e) => setSelectedBiller(e.target.value)}
          >
            {POPULAR_BILLERS.map((b) => (
              <option key={b.id} value={b.name}>
                {b.name} ({b.category})
              </option>
            ))}
            <option value="custom">+ Other / Custom Biller</option>
          </select>
        </div>

        {selectedBiller === 'custom' && (
          <div className="form-group">
            <label htmlFor="custom-biller">Biller Name</label>
            <input
              id="custom-biller"
              type="text"
              placeholder="e.g. City Gas Co."
              value={customBiller}
              onChange={(e) => setCustomBiller(e.target.value)}
              required
            />
          </div>
        )}

        <div className="form-group">
          <label htmlFor="bill-reference">
            Account / Invoice Reference Number
            <small>Found on your monthly bill</small>
          </label>
          <input
            id="bill-reference"
            type="text"
            placeholder="e.g. ACC-98214-X"
            value={reference}
            onChange={(e) => setReference(e.target.value)}
            required
          />
        </div>

        <div className="form-group">
          <label htmlFor="bill-amount">
            Amount (whole dollars, min $1.00)
            <small>No decimals</small>
          </label>
          <input
            id="bill-amount"
            type="number"
            min="1"
            max="1000000"
            step="1"
            placeholder="e.g. 75"
            value={amount}
            onKeyDown={(e) => {
              if (BLOCKED_KEYS.includes(e.key)) e.preventDefault();
            }}
            onChange={(e) => setAmount(e.target.value)}
            required
          />

          <div className="quick-amounts">
            {QUICK_AMOUNTS.map((q) => (
              <button
                type="button"
                key={q}
                onClick={() => setAmount(String(q))}
              >
                ${q}
              </button>
            ))}
          </div>
        </div>

        <button type="submit" className="btn btn-primary" disabled={loading}>
          {loading ? 'Processing Payment...' : 'Pay Bill'}
        </button>
      </form>
    </section>
  );
}
