import { useState } from 'react';
import { api } from '../../api/client.js';
import { formatCardNumber, formatDate, formatIBAN } from '../../utils/format.js';
import AccountQR from './AccountQR.jsx';
import DebitCard from './DebitCard.jsx';

export default function ManageAccountPanel({ user, onUserUpdated, onNotify }) {
  const [editingName, setEditingName] = useState(false);
  const [nameInput, setNameInput] = useState(user?.name || '');
  const [nameLoading, setNameLoading] = useState(false);
  const [editingCard, setEditingCard] = useState(false);
  const [cardForm, setCardForm] = useState({
    card_number: user?.card_number || '',
    card_expiry: user?.card_expiry || '',
    iban: user?.iban || '',
  });
  const [cardLoading, setCardLoading] = useState(false);

  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [passwordLoading, setPasswordLoading] = useState(false);
  const [passwordMsg, setPasswordMsg] = useState({ text: '', type: 'error' });

  const [copiedAcct, setCopiedAcct] = useState(false);
  const [showCardNumber, setShowCardNumber] = useState(false);
  const [copiedCard, setCopiedCard] = useState(false);
  const [copiedIBAN, setCopiedIBAN] = useState(false);

  function copyAccountNumber() {
    if (!user?.account_number) return;
    navigator.clipboard.writeText(user.account_number);
    setCopiedAcct(true);
    setTimeout(() => setCopiedAcct(false), 2000);
  }

  function copyCardNumber() {
    if (!user?.card_number) return;
    navigator.clipboard.writeText(user.card_number);
    setCopiedCard(true);
    setTimeout(() => setCopiedCard(false), 2000);
  }

  function copyIBAN() {
    if (!user?.iban) return;
    navigator.clipboard.writeText(user.iban);
    setCopiedIBAN(true);
    setTimeout(() => setCopiedIBAN(false), 2000);
  }

  async function handleUpdateName(e) {
    e.preventDefault();
    const trimmed = nameInput.trim();
    if (trimmed.length < 2 || trimmed.length > 60) {
      onNotify('Name must be between 2 and 60 characters.', 'error');
      return;
    }
    setNameLoading(true);
    try {
      const updated = await api.updateProfile(trimmed);
      onUserUpdated(updated);
      setEditingName(false);
      onNotify('Profile name updated successfully.', 'success');
    } catch (err) {
      onNotify(err.message, 'error');
    } finally {
      setNameLoading(false);
    }
  }

  async function handleUpdatePassword(e) {
    e.preventDefault();
    setPasswordMsg({ text: '', type: 'error' });
    if (!currentPassword) {
      setPasswordMsg({ text: 'Please enter your current password.', type: 'error' });
      return;
    }
    if (newPassword.length < 8 || newPassword.length > 72) {
      setPasswordMsg({ text: 'New password must be between 8 and 72 characters.', type: 'error' });
      return;
    }
    if (newPassword !== confirmPassword) {
      setPasswordMsg({ text: 'New passwords do not match.', type: 'error' });
      return;
    }

    setPasswordLoading(true);
    try {
      await api.updatePassword(currentPassword, newPassword);
      setPasswordMsg({ text: 'Password changed successfully.', type: 'success' });
      setCurrentPassword('');
      setNewPassword('');
      setConfirmPassword('');
    } catch (err) {
      setPasswordMsg({ text: err.message, type: 'error' });
    } finally {
      setPasswordLoading(false);
    }
  }

  async function handleUpdateCard(e) {
    e.preventDefault();
    setCardLoading(true);
    try {
      const updated = await api.updateCard(cardForm.card_number, cardForm.card_expiry, cardForm.iban);
      onUserUpdated(updated);
      setEditingCard(false);
      onNotify('Card and banking details updated successfully.', 'success');
    } catch (err) {
      onNotify(err.message, 'error');
    } finally {
      setCardLoading(false);
    }
  }

  return (
    <div className="manage-account-wrapper">
      {/* Profile Debit Card Showcase */}
      <div className="card profile-card-showcase">
        <h2>Your Virtual Debit Card &amp; Banking Details</h2>
        <p className="muted" style={{ marginBottom: '1.25rem' }}>
          Use your 16-digit card number and IBAN for seamless domestic and international operations.
        </p>
        <DebitCard user={user} />
        {editingCard ? (
          <form className="form" onSubmit={handleUpdateCard}>
            <label>
              Card number
              <input
                type="text"
                inputMode="numeric"
                autoComplete="cc-number"
                maxLength={23}
                value={cardForm.card_number}
                onChange={(e) => setCardForm({ ...cardForm, card_number: e.target.value })}
                required
              />
            </label>
            <label>
              Card expiry <small>(MM/YY)</small>
              <input
                type="text"
                inputMode="numeric"
                autoComplete="cc-exp"
                placeholder="MM/YY"
                maxLength={5}
                value={cardForm.card_expiry}
                onChange={(e) => setCardForm({ ...cardForm, card_expiry: e.target.value })}
                required
              />
            </label>
            <label>
              IBAN
              <input
                type="text"
                autoComplete="off"
                maxLength={34}
                value={cardForm.iban}
                onChange={(e) => setCardForm({ ...cardForm, iban: e.target.value })}
                required
              />
            </label>
            <div className="detail-value-row">
              <button type="submit" className="btn btn-primary" disabled={cardLoading}>
                {cardLoading ? 'Saving...' : 'Save banking details'}
              </button>
              <button
                type="button"
                className="btn btn-ghost-dark"
                onClick={() => setEditingCard(false)}
                disabled={cardLoading}
              >
                Cancel
              </button>
            </div>
          </form>
        ) : (
          <button
            type="button"
            className="btn btn-secondary"
            onClick={() => {
              setCardForm({
                card_number: user?.card_number || '',
                card_expiry: user?.card_expiry || '',
                iban: user?.iban || '',
              });
              setEditingCard(true);
            }}
          >
            Edit card &amp; banking details
          </button>
        )}
      </div>

      <section className="manage-account-grid">
        {/* Account Details Card */}
        <div className="card">
          <h2>Account Profile</h2>
          <div className="account-details-list">
            <div className="detail-item">
              <span className="detail-label">Full Name</span>
              {editingName ? (
                <form className="inline-edit-form" onSubmit={handleUpdateName}>
                  <input
                    type="text"
                    value={nameInput}
                    onChange={(e) => setNameInput(e.target.value)}
                    disabled={nameLoading}
                    required
                  />
                  <button type="submit" className="btn btn-primary btn-sm" disabled={nameLoading}>
                    Save
                  </button>
                  <button
                    type="button"
                    className="btn btn-ghost-dark btn-sm"
                    onClick={() => {
                      setNameInput(user.name);
                      setEditingName(false);
                    }}
                    disabled={nameLoading}
                  >
                    Cancel
                  </button>
                </form>
              ) : (
                <div className="detail-value-row">
                  <strong className="detail-value">{user.name}</strong>
                  <button
                    type="button"
                    className="text-link-btn"
                    onClick={() => {
                      setNameInput(user.name);
                      setEditingName(true);
                    }}
                  >
                    Edit
                  </button>
                </div>
              )}
            </div>

            <div className="detail-item">
              <span className="detail-label">Card Number</span>
              <div className="detail-value-row">
                <strong className="detail-value monospace">
                  {formatCardNumber(user?.card_number, !showCardNumber) || '5342 •••• •••• ••••'}
                </strong>
                <div style={{ display: 'flex', gap: '0.4rem' }}>
                  <button
                    type="button"
                    className="btn btn-ghost-dark btn-sm"
                    onClick={() => setShowCardNumber((prev) => !prev)}
                  >
                    {showCardNumber ? 'Hide' : 'Show'}
                  </button>
                  <button
                    type="button"
                    className="btn btn-ghost-dark btn-sm"
                    onClick={copyCardNumber}
                  >
                    {copiedCard ? '✓ Copied!' : 'Copy'}
                  </button>
                </div>
              </div>
            </div>

            <div className="detail-item">
              <span className="detail-label">Date of Card (Valid Thru)</span>
              <div className="detail-value-row">
                <strong className="detail-value monospace">{user?.card_expiry || '12/30'}</strong>
                <span className="status-badge active">&#9679; Active</span>
              </div>
            </div>

            <div className="detail-item">
              <span className="detail-label">IBAN (International Bank Account)</span>
              <div className="detail-value-row">
                <strong className="detail-value monospace">{formatIBAN(user?.iban) || user?.iban || '—'}</strong>
                <button
                  type="button"
                  className="btn btn-ghost-dark btn-sm"
                  onClick={copyIBAN}
                >
                  {copiedIBAN ? '✓ Copied!' : 'Copy'}
                </button>
              </div>
            </div>

            <div className="detail-item">
              <span className="detail-label">Account Number</span>
              <div className="detail-value-row">
                <strong className="detail-value monospace">{user.account_number}</strong>
                <button
                  type="button"
                  className="btn btn-ghost-dark btn-sm"
                  onClick={copyAccountNumber}
                >
                  {copiedAcct ? '✓ Copied!' : 'Copy'}
                </button>
              </div>
            </div>

            <div className="detail-item">
              <span className="detail-label">Email Address</span>
              <span className="detail-value">{user.email}</span>
            </div>

            <div className="detail-item">
              <span className="detail-label">Account Status</span>
              <span className="status-badge active">&#9679; Active &amp; Verified</span>
            </div>

            {user.created_at && (
              <div className="detail-item">
                <span className="detail-label">Member Since</span>
                <span className="detail-value">{formatDate(user.created_at)}</span>
              </div>
            )}
          </div>
        </div>

        {/* Account Unique QR Code Card */}
        <div className="card">
          <h2>Account QR Code</h2>
          <p className="muted" style={{ marginBottom: '1rem' }}>
            Your unique account QR code for fast, touch-free peer transfers and payment requests.
          </p>

          <AccountQR user={user} size={190} showDetails={true} />

          <div className="qr-explanation">
            <p>
              Other MIeBank users can scan or upload this QR code to instantly send money to your account without typing your account number.
            </p>
          </div>
        </div>

        {/* Security & Password Card */}
        <div className="card">
          <h2>Security &amp; Password</h2>
          <p className="muted" style={{ marginBottom: '1rem' }}>
            Update your login password to keep your account safe.
          </p>

          {passwordMsg.text && (
            <div className={`message ${passwordMsg.type}`} style={{ marginBottom: '1rem' }}>
              {passwordMsg.text}
            </div>
          )}

          <form className="form inline-form" onSubmit={handleUpdatePassword}>
            <div className="form-group">
              <label htmlFor="current-password">Current Password</label>
              <input
                id="current-password"
                type="password"
                value={currentPassword}
                onChange={(e) => setCurrentPassword(e.target.value)}
                required
              />
            </div>

            <div className="form-group">
              <label htmlFor="new-password">New Password (min 8 characters)</label>
              <input
                id="new-password"
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                required
              />
            </div>

            <div className="form-group">
              <label htmlFor="confirm-password">Confirm New Password</label>
              <input
                id="confirm-password"
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                required
              />
            </div>

            <button type="submit" className="btn btn-primary" disabled={passwordLoading}>
              {passwordLoading ? 'Updating Password...' : 'Update Password'}
            </button>
          </form>
        </div>

        {/* Banking Limits Card */}
        <div className="card">
          <h2>Account Limits &amp; Policies</h2>
          <ul className="limits-list">
            <li>
              <span>Minimum deposit &amp; withdrawal:</span>
              <strong>$20.00</strong>
            </li>
            <li>
              <span>Minimum transfer / bill / donation:</span>
              <strong>$1.00</strong>
            </li>
            <li>
              <span>Maximum per-transaction limit:</span>
              <strong>$1,000,000.00</strong>
            </li>
            <li>
              <span>Overdraft Protection:</span>
              <strong>Standard (Zero-balance floor)</strong>
            </li>
            <li>
              <span>Internal Transfer Speed:</span>
              <strong className="accent">Instant (0 seconds)</strong>
            </li>
            <li>
              <span>Monthly Account Maintenance Fee:</span>
              <strong>$0.00 (Free)</strong>
            </li>
          </ul>
        </div>
      </section>
    </div>
  );
}
