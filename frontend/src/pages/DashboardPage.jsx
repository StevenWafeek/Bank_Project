import { useCallback, useEffect, useState } from 'react';
import { api } from '../api/client.js';
import { useAuth } from '../context/AuthContext.jsx';
import { formatMoney } from '../utils/format.js';
import OptionCards from '../components/dashboard/OptionCards.jsx';
import BalancePanel from '../components/dashboard/BalancePanel.jsx';
import MoneyForm from '../components/dashboard/MoneyForm.jsx';
import TransferPanel from '../components/dashboard/TransferPanel.jsx';
import BillsPanel from '../components/dashboard/BillsPanel.jsx';
import DonatePanel from '../components/dashboard/DonatePanel.jsx';
import ManageAccountPanel from '../components/dashboard/ManageAccountPanel.jsx';
import TransactionsTable from '../components/dashboard/TransactionsTable.jsx';
import Message from '../components/Message.jsx';

export default function DashboardPage() {
  const { user, setUser, handleAuthError } = useAuth();
  const [panel, setPanel] = useState('balance');
  const [balanceCents, setBalanceCents] = useState(0);
  const [transactions, setTransactions] = useState([]);
  const [message, setMessage] = useState({ text: '', type: 'error' });

  const refresh = useCallback(async () => {
    try {
      const [bal, txs] = await Promise.all([api.balance(), api.transactions()]);
      setBalanceCents(bal.balance_cents);
      setTransactions(txs.transactions);
    } catch (err) {
      if (!handleAuthError(err)) setMessage({ text: err.message, type: 'error' });
    }
  }, [handleAuthError]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  function selectPanel(next) {
    setPanel(next);
    setMessage({ text: '', type: 'error' });
    if (next === 'balance') refresh();
  }

  // Returns true when the transaction succeeded so the form can clear itself.
  async function handleTransaction(type, amount, validationError) {
    if (validationError) {
      setMessage({ text: validationError, type: 'error' });
      return false;
    }
    try {
      const res = await (type === 'deposit' ? api.deposit(amount) : api.withdraw(amount));
      const verb = type === 'deposit' ? 'Deposited' : 'Withdrew';
      setMessage({
        text: `${verb} ${formatMoney(res.transaction.amount_cents)} successfully. New balance: ${formatMoney(res.balance_cents)}`,
        type: 'success',
      });
      await refresh();
      return true;
    } catch (err) {
      if (!handleAuthError(err)) setMessage({ text: err.message, type: 'error' });
      return false;
    }
  }

  async function handleTransfer(validationError, data) {
    if (validationError) {
      setMessage({ text: validationError, type: 'error' });
      return false;
    }
    try {
      const res = await api.transfer(data.recipient, data.amount, data.note);
      setMessage({
        text: `Transferred ${formatMoney(res.transaction.amount_cents)} to ${res.recipient_name} (${res.recipient_account}) successfully. New balance: ${formatMoney(res.balance_cents)}`,
        type: 'success',
      });
      await refresh();
      return true;
    } catch (err) {
      if (!handleAuthError(err)) setMessage({ text: err.message, type: 'error' });
      return false;
    }
  }

  async function handlePayBill(validationError, data) {
    if (validationError) {
      setMessage({ text: validationError, type: 'error' });
      return false;
    }
    try {
      const res = await api.payBill(data.biller, data.reference, data.amount);
      setMessage({
        text: `Paid ${formatMoney(res.transaction.amount_cents)} for ${res.biller} (Ref: ${res.reference}) successfully! Confirmation: ${res.confirmation_code}. New balance: ${formatMoney(res.balance_cents)}`,
        type: 'success',
      });
      await refresh();
      return true;
    } catch (err) {
      if (!handleAuthError(err)) setMessage({ text: err.message, type: 'error' });
      return false;
    }
  }

  async function handleDonate(validationError, data) {
    if (validationError) {
      setMessage({ text: validationError, type: 'error' });
      return false;
    }
    try {
      const res = await api.donate(data.charity, data.amount, data.note, data.anonymous);
      setMessage({
        text: `Donated ${formatMoney(res.transaction.amount_cents)} to ${res.charity} successfully! Receipt ID: ${res.receipt_id}. Thank you for your generosity!`,
        type: 'success',
      });
      await refresh();
      return true;
    } catch (err) {
      if (!handleAuthError(err)) setMessage({ text: err.message, type: 'error' });
      return false;
    }
  }

  return (
    <main className="container">
      <section className="welcome">
        <h1>
          Hello, <span id="user-name">{user.name}</span>
        </h1>
        <p className="muted">
          Account <strong id="account-number">{user.account_number}</strong>
        </p>
      </section>

      <OptionCards active={panel} onSelect={selectPanel} />

      {panel === 'balance' && (
        <BalancePanel
          balanceCents={balanceCents}
          user={user}
          onAction={selectPanel}
        />
      )}

      {panel === 'transfer' && (
        <TransferPanel
          balanceCents={balanceCents}
          onSubmit={handleTransfer}
        />
      )}

      {(panel === 'deposit' || panel === 'withdraw') && (
        <MoneyForm
          key={panel}
          type={panel}
          balanceCents={balanceCents}
          onSubmit={handleTransaction}
        />
      )}

      {panel === 'bills' && (
        <BillsPanel
          balanceCents={balanceCents}
          onSubmit={handlePayBill}
        />
      )}

      {panel === 'donate' && (
        <DonatePanel
          balanceCents={balanceCents}
          onSubmit={handleDonate}
        />
      )}

      {panel === 'manage' && (
        <ManageAccountPanel
          user={user}
          onUserUpdated={(updated) => setUser({ ...user, ...updated })}
          onNotify={(text, type) => setMessage({ text, type })}
        />
      )}

      <Message id="dash-message" text={message.text} type={message.type} />

      <TransactionsTable transactions={transactions} />
    </main>
  );
}
