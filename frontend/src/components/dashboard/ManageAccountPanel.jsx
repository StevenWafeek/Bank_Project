import { useState } from 'react';
import { api } from '../../api/client.js';
import { formatDate } from '../../utils/format.js';
import AccountQR from './AccountQR.jsx';

export default function ManageAccountPanel({ user, onUserUpdated, onNotify }) {
  const [editingName, setEditingName] = useState(false);
  const [nameInput, setNameInput] = useState(user?.name || '');
  const [nameLoading, setNameLoading] = useState(false);

  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [passwordLoading, setPasswordLoading] = useState(false);
  const [passwordMsg, setPasswordMsg] = useState({ text: '', type: 'error' });

  const [copiedAcct, setCopiedAcct] = useState(false);

  function copyAccountNumber() {
    if (!user?.account_number) return;
    navigator.clipboard.writeText(user.account_number);
    setCopiedAcct(true);
    setTimeout(() => setCopiedAcct(false), 2000);
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

  return (
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
            <span className="status-badge active">&#9679; Active & Verified</span>
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
        <h2>Security & Password</h2>
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
        <h2>Account Limits & Policies</h2>
        <ul className="limits-list">
          <li>
            <span>Minimum deposit & withdrawal:</span>
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
  );
}
