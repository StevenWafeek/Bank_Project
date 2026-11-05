import { useState } from 'react';
import { formatMoney } from '../../utils/format.js';
import AccountQR from './AccountQR.jsx';

export default function BalancePanel({ balanceCents, user, onAction }) {
  const [showQR, setShowQR] = useState(false);

  return (
    <section className="card panel" id="panel-balance">
      <div className="balance-header">
        <div>
          <h2>Your Balance</h2>
          <p className="balance" id="balance-amount">
            {formatMoney(balanceCents)}
          </p>
          <p className="muted">Available balance &bull; Account: <strong>{user?.account_number}</strong></p>
        </div>

        <button
          type="button"
          className="btn btn-secondary qr-toggle-btn"
          onClick={() => setShowQR((prev) => !prev)}
        >
          <svg
            width="18"
            height="18"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
            style={{ marginRight: '6px' }}
          >
            <rect x="3" y="3" width="7" height="7" />
            <rect x="14" y="3" width="7" height="7" />
            <rect x="14" y="14" width="7" height="7" />
            <rect x="3" y="14" width="7" height="7" />
          </svg>
          {showQR ? 'Hide My QR' : 'Show My QR'}
        </button>
      </div>

      {showQR && (
        <div className="balance-qr-dropdown">
          <p className="qr-hint">Scan with another account to receive money instantly:</p>
          <AccountQR user={user} size={160} />
        </div>
      )}

      {onAction && (
        <div className="quick-actions-bar">
          <button type="button" className="action-pill" onClick={() => onAction('transfer')}>
            &rarr; Transfer
          </button>
          <button type="button" className="action-pill" onClick={() => onAction('deposit')}>
            &darr; Deposit
          </button>
          <button type="button" className="action-pill" onClick={() => onAction('withdraw')}>
            &uarr; Withdraw
          </button>
          <button type="button" className="action-pill" onClick={() => onAction('bills')}>
            &#9776; Pay Bills
          </button>
        </div>
      )}
    </section>
  );
}
