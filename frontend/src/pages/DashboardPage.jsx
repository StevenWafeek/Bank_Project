import { useCallback, useEffect, useState } from 'react';
import { api } from '../api/client.js';
import { useAuth } from '../context/AuthContext.jsx';
import { formatMoney } from '../utils/format.js';
import OptionCards from '../components/dashboard/OptionCards.jsx';
import BalancePanel from '../components/dashboard/BalancePanel.jsx';
import MoneyForm from '../components/dashboard/MoneyForm.jsx';
import TransactionsTable from '../components/dashboard/TransactionsTable.jsx';
import Message from '../components/Message.jsx';

export default function DashboardPage() {
  const { user, handleAuthError } = useAuth();
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

      {panel === 'balance' ? (
        <BalancePanel balanceCents={balanceCents} />
      ) : (
        <MoneyForm key={panel} type={panel} balanceCents={balanceCents} onSubmit={handleTransaction} />
      )}

      <Message id="dash-message" text={message.text} type={message.type} />

      <TransactionsTable transactions={transactions} />
    </main>
  );
}
