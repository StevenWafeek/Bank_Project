import { useState } from 'react';
import { api } from '../../api/client.js';
import { BLOCKED_KEYS, validatePaymentAmount } from '../../utils/amount.js';
import { formatMoney } from '../../utils/format.js';

const QUICK_AMOUNTS = [20, 50, 100, 250, 500];

export default function TransferPanel({ balanceCents, onSubmit }) {
  const [recipient, setRecipient] = useState('');
  const [amount, setAmount] = useState('');
  const [note, setNote] = useState('');
  const [verifiedName, setVerifiedName] = useState(null);
  const [lookupError, setLookupError] = useState('');
  const [isLookingUp, setIsLookingUp] = useState(false);
  const [showQRPaste, setShowQRPaste] = useState(false);
  const [qrRawInput, setQrRawInput] = useState('');
  const [loading, setLoading] = useState(false);

  async function handleVerify(queryToVerify) {
    const q = (queryToVerify || recipient).trim();
    if (!q) {
      setVerifiedName(null);
      setLookupError('');
      return;
    }
    setIsLookingUp(true);
    setLookupError('');
    try {
      const res = await api.lookupRecipient(q);
      setVerifiedName(res.name);
      setRecipient(res.account_number);
    } catch (err) {
      setVerifiedName(null);
      setLookupError(err.message || 'Recipient not found');
    } finally {
      setIsLookingUp(false);
    }
  }

  function handleParseQR() {
    const raw = qrRawInput.trim();
    if (!raw) return;
    try {
      // Try JSON
      if (raw.startsWith('{')) {
        const parsed = JSON.parse(raw);
        const acct = parsed.account_number || parsed.account || '';
        if (acct) {
          setRecipient(acct);
          if (parsed.name) setVerifiedName(parsed.name);
          setShowQRPaste(false);
          setQrRawInput('');
          handleVerify(acct);
          return;
        }
      }
      // Try URI: miebank://... or query
      if (raw.includes('account=')) {
        const match = raw.match(/account=([A-Za-z0-9]+)/);
        if (match && match[1]) {
          setRecipient(match[1]);
          setShowQRPaste(false);
          setQrRawInput('');
          handleVerify(match[1]);
          return;
        }
      }
      // Raw string
      setRecipient(raw);
      setShowQRPaste(false);
      setQrRawInput('');
      handleVerify(raw);
    } catch {
      setLookupError('Could not parse QR code data');
    }
  }

  async function handleSubmit(e) {
    e.preventDefault();
    const err = validatePaymentAmount(amount, 1);
    if (err) {
      onSubmit(err);
      return;
    }
    const cents = Number(amount) * 100;
    if (cents > balanceCents) {
      onSubmit('Insufficient funds for this transfer.');
      return;
    }
    if (!recipient.trim()) {
      onSubmit('Please enter a recipient account number or email.');
      return;
    }

    setLoading(true);
    const ok = await onSubmit(null, {
      recipient: recipient.trim(),
      amount: Number(amount),
      note: note.trim(),
    });
    setLoading(false);
    if (ok) {
      setAmount('');
      setNote('');
      setRecipient('');
      setVerifiedName(null);
    }
  }

  return (
    <section className="card panel" id="panel-transfer">
      <h2>Transfer to Another Account</h2>
      <p className="muted">
        Send funds instantly to any MIeBank user. Available balance:{' '}
        <strong className="accent">{formatMoney(balanceCents)}</strong>
      </p>

      <form className="form inline-form" onSubmit={handleSubmit}>
        <div className="form-group">
          <div className="label-with-action">
            <label htmlFor="transfer-recipient">Recipient Account or Email</label>
            <button
              type="button"
              className="text-link-btn"
              onClick={() => setShowQRPaste((prev) => !prev)}
            >
              {showQRPaste ? 'Cancel QR' : '📷 Paste QR Code Data'}
            </button>
          </div>

          {showQRPaste && (
            <div className="qr-paste-box">
              <label htmlFor="qr-paste-input">Paste QR Code JSON or payload:</label>
              <textarea
                id="qr-paste-input"
                className="qr-textarea"
                rows="2"
                placeholder='e.g. {"account_number":"MIE0000000002"}'
                value={qrRawInput}
                onChange={(e) => setQrRawInput(e.target.value)}
              />
              <button
                type="button"
                className="btn btn-secondary btn-sm"
                onClick={handleParseQR}
              >
                Apply QR Payload
              </button>
            </div>
          )}

          <div className="input-with-button">
            <input
              id="transfer-recipient"
              type="text"
              placeholder="e.g. MIE0123456789 or user@bank.com"
              value={recipient}
              onChange={(e) => {
                setRecipient(e.target.value);
                setVerifiedName(null);
                setLookupError('');
              }}
              onBlur={() => handleVerify()}
              required
            />
            <button
              type="button"
              className="btn btn-secondary"
              onClick={() => handleVerify()}
              disabled={isLookingUp || !recipient.trim()}
            >
              {isLookingUp ? 'Checking...' : 'Verify'}
            </button>
          </div>

          {verifiedName && (
            <div className="verified-badge">
              <span className="check-icon">&#10003;</span> Recipient verified: <strong>{verifiedName}</strong>
            </div>
          )}
          {lookupError && <div className="field-error">{lookupError}</div>}
        </div>

        <div className="form-group">
          <label htmlFor="transfer-amount">
            Amount (whole dollars, min $1.00)
            <small>No decimals or coins</small>
          </label>
          <input
            id="transfer-amount"
            type="number"
            min="1"
            max="1000000"
            step="1"
            placeholder="e.g. 50"
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

        <div className="form-group">
          <label htmlFor="transfer-note">
            Note / Memo <small>(Optional)</small>
          </label>
          <input
            id="transfer-note"
            type="text"
            maxLength="100"
            placeholder="e.g. Dinner split, Rent, Birthday gift"
            value={note}
            onChange={(e) => setNote(e.target.value)}
          />
        </div>

        <button type="submit" className="btn btn-primary" disabled={loading}>
          {loading ? 'Processing Transfer...' : 'Send Transfer'}
        </button>
      </form>
    </section>
  );
}
