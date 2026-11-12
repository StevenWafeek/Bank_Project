import { useState } from 'react';
import { formatCardNumber, formatIBAN } from '../../utils/format.js';

export default function DebitCard({ user, compact = false }) {
  const [showFullNumber, setShowFullNumber] = useState(false);
  const [copiedCard, setCopiedCard] = useState(false);
  const [copiedIBAN, setCopiedIBAN] = useState(false);
  const [copiedAccount, setCopiedAccount] = useState(false);

  const cardNumber = user?.card_number || '';
  const cardExpiry = user?.card_expiry || '12/30';
  const iban = user?.iban || '';
  const accountNumber = user?.account_number || '';
  const holderName = (user?.name || 'CARD HOLDER').toUpperCase();

  function copyCardNumber() {
    if (!cardNumber) return;
    navigator.clipboard.writeText(cardNumber);
    setCopiedCard(true);
    setTimeout(() => setCopiedCard(false), 2000);
  }

  function copyIBAN() {
    if (!iban) return;
    navigator.clipboard.writeText(iban);
    setCopiedIBAN(true);
    setTimeout(() => setCopiedIBAN(false), 2000);
  }

  function copyAccount() {
    if (!accountNumber) return;
    navigator.clipboard.writeText(accountNumber);
    setCopiedAccount(true);
    setTimeout(() => setCopiedAccount(false), 2000);
  }

  const displayedCardNumber = formatCardNumber(cardNumber, !showFullNumber);

  return (
    <div className={`debit-card-container ${compact ? 'compact' : ''}`}>
      {/* Physical Virtual Card Graphic */}
      <div className="debit-card">
        <div className="card-top-row">
          <div className="card-brand">
            <span className="card-bank-name">MIeBank</span>
            <span className="card-badge">DEBIT</span>
          </div>

          <div className="card-contactless">
            {/* Contactless waves SVG */}
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round">
              <path d="M8.5 16.5a5 5 0 0 1 0-9" />
              <path d="M12 19a8.5 8.5 0 0 0 0-14" />
              <path d="M15.5 21.5a12 12 0 0 0 0-19" />
            </svg>
          </div>
        </div>

        {/* Golden EMV Smart Chip */}
        <div className="card-chip-row">
          <div className="emv-chip" title="EMV Smart Chip">
            <div className="chip-line horizontal"></div>
            <div className="chip-line vertical"></div>
          </div>
        </div>

        {/* 16-Digit Card Number */}
        <div className="card-number-row">
          <div className="card-number" id="debit-card-number">
            {displayedCardNumber || '5342 •••• •••• ••••'}
          </div>

          <div className="card-number-actions">
            <button
              type="button"
              className="card-icon-btn"
              onClick={() => setShowFullNumber((prev) => !prev)}
              title={showFullNumber ? 'Hide digits' : 'Show full card number'}
              aria-label={showFullNumber ? 'Hide digits' : 'Show full card number'}
            >
              {showFullNumber ? (
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24" />
                  <line x1="1" y1="1" x2="23" y2="23" />
                </svg>
              ) : (
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
                  <circle cx="12" cy="12" r="3" />
                </svg>
              )}
            </button>

            <button
              type="button"
              className="card-icon-btn"
              onClick={copyCardNumber}
              title="Copy card number"
              aria-label="Copy card number"
            >
              {copiedCard ? (
                <span className="card-btn-text">✓ Copied</span>
              ) : (
                <svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
                  <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
                </svg>
              )}
            </button>
          </div>
        </div>

        {/* Cardholder Name & Expiry Date */}
        <div className="card-bottom-row">
          <div className="card-holder-info">
            <span className="card-label">CARD HOLDER</span>
            <span className="card-holder-name">{holderName}</span>
          </div>

          <div className="card-expiry-info">
            <span className="card-label">VALID THRU</span>
            <span className="card-expiry-date">{cardExpiry}</span>
          </div>

          <div className="card-brand-circles">
            <span className="circle circle-red"></span>
            <span className="circle circle-yellow"></span>
          </div>
        </div>
      </div>

      {/* IBAN & Account Quick Bar */}
      <div className="card-meta-bar">
        <div className="meta-bar-item">
          <span className="meta-bar-label">IBAN</span>
          <strong className="meta-bar-val monospace">{formatIBAN(iban) || iban || '—'}</strong>
          <button
            type="button"
            className="btn btn-ghost-dark btn-sm meta-copy-btn"
            onClick={copyIBAN}
            title="Copy IBAN to clipboard"
          >
            {copiedIBAN ? '✓ Copied' : 'Copy IBAN'}
          </button>
        </div>

        <div className="meta-bar-item">
          <span className="meta-bar-label">Account Number</span>
          <strong className="meta-bar-val monospace">{accountNumber || '—'}</strong>
          <button
            type="button"
            className="btn btn-ghost-dark btn-sm meta-copy-btn"
            onClick={copyAccount}
            title="Copy Account Number"
          >
            {copiedAccount ? '✓ Copied' : 'Copy'}
          </button>
        </div>
      </div>
    </div>
  );
}
