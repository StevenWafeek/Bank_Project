import { useState } from 'react';
import { BLOCKED_KEYS, validatePaymentAmount } from '../../utils/amount.js';
import { formatMoney } from '../../utils/format.js';

const CHARITIES = [
  {
    id: 'redcross',
    name: 'Red Cross & Disaster Relief',
    desc: 'Emergency assistance, disaster preparedness, and blood drives',
  },
  {
    id: 'stjude',
    name: "St. Jude Children's Research Hospital",
    desc: 'Treating childhood cancer and pediatric catastrophic diseases',
  },
  {
    id: 'food',
    name: 'Global Food Program & Food Banks',
    desc: 'Nutritious meals for families and communities in crisis',
  },
  {
    id: 'water',
    name: 'Clean Water Initiative',
    desc: 'Sustainable clean drinking water wells for underserved regions',
  },
  {
    id: 'education',
    name: 'Education For All Fund',
    desc: 'Books, STEM labs, and tuition scholarships for youth',
  },
  {
    id: 'wwf',
    name: 'World Wildlife & Conservation Fund',
    desc: 'Protecting endangered species and fragile natural ecosystems',
  },
];

const DONATION_CHIPS = [10, 25, 50, 100, 250];

export default function DonatePanel({ balanceCents, onSubmit }) {
  const [selectedCharity, setSelectedCharity] = useState(CHARITIES[0].name);
  const [customCharity, setCustomCharity] = useState('');
  const [amount, setAmount] = useState('25');
  const [note, setNote] = useState('');
  const [anonymous, setAnonymous] = useState(false);
  const [loading, setLoading] = useState(false);

  const charityName = selectedCharity === 'custom' ? customCharity : selectedCharity;
  const activeCharityObj = CHARITIES.find((c) => c.name === selectedCharity);

  async function handleSubmit(e) {
    e.preventDefault();
    if (!charityName.trim()) {
      onSubmit('Please enter a charity name.');
      return;
    }
    const err = validatePaymentAmount(amount, 1);
    if (err) {
      onSubmit(err);
      return;
    }
    const cents = Number(amount) * 100;
    if (cents > balanceCents) {
      onSubmit('Insufficient funds for this donation.');
      return;
    }

    setLoading(true);
    const ok = await onSubmit(null, {
      charity: charityName.trim(),
      amount: Number(amount),
      note: note.trim(),
      anonymous,
    });
    setLoading(false);
    if (ok) {
      setAmount('25');
      setNote('');
      if (selectedCharity === 'custom') setCustomCharity('');
    }
  }

  return (
    <section className="card panel" id="panel-donate">
      <h2>Charity Donations</h2>
      <p className="muted">
        Make a positive difference in the world. Available balance:{' '}
        <strong className="accent">{formatMoney(balanceCents)}</strong>
      </p>

      <form className="form inline-form" onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="charity-select">Select Organization</label>
          <select
            id="charity-select"
            className="form-select"
            value={selectedCharity}
            onChange={(e) => setSelectedCharity(e.target.value)}
          >
            {CHARITIES.map((c) => (
              <option key={c.id} value={c.name}>
                {c.name}
              </option>
            ))}
            <option value="custom">+ Other Charitable Organization</option>
          </select>

          {activeCharityObj && (
            <div className="charity-desc-callout">
              <span className="heart-icon">&hearts;</span> {activeCharityObj.desc}
            </div>
          )}
        </div>

        {selectedCharity === 'custom' && (
          <div className="form-group">
            <label htmlFor="custom-charity">Organization Name</label>
            <input
              id="custom-charity"
              type="text"
              placeholder="e.g. Local Animal Shelter"
              value={customCharity}
              onChange={(e) => setCustomCharity(e.target.value)}
              required
            />
          </div>
        )}

        <div className="form-group">
          <label htmlFor="donation-amount">
            Donation Amount (whole dollars, min $1.00)
            <small>All donations receive an official tax receipt reference</small>
          </label>
          <input
            id="donation-amount"
            type="number"
            min="1"
            max="1000000"
            step="1"
            placeholder="e.g. 25"
            value={amount}
            onKeyDown={(e) => {
              if (BLOCKED_KEYS.includes(e.key)) e.preventDefault();
            }}
            onChange={(e) => setAmount(e.target.value)}
            required
          />

          <div className="quick-amounts">
            {DONATION_CHIPS.map((chip) => (
              <button
                type="button"
                key={chip}
                className={amount === String(chip) ? 'active' : ''}
                onClick={() => setAmount(String(chip))}
              >
                ${chip}
              </button>
            ))}
          </div>
        </div>

        <div className="form-group">
          <label htmlFor="donation-note">
            Dedication / Personal Message <small>(Optional)</small>
          </label>
          <input
            id="donation-note"
            type="text"
            placeholder="e.g. In memory of..., Keep inspiring!"
            value={note}
            onChange={(e) => setNote(e.target.value)}
          />
        </div>

        <div className="checkbox-row">
          <label className="checkbox-label">
            <input
              type="checkbox"
              checked={anonymous}
              onChange={(e) => setAnonymous(e.target.checked)}
            />
            <span>Make this donation anonymous</span>
          </label>
        </div>

        <button type="submit" className="btn btn-success" disabled={loading}>
          {loading ? 'Processing Donation...' : 'Donate Now'}
        </button>
      </form>
    </section>
  );
}
