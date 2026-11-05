import { useState } from 'react';
import { formatDate, formatMoney } from '../../utils/format.js';

const TYPE_CONFIG = {
  deposit: { label: 'Deposit', sign: '+', isPos: true, badgeClass: 'badge-deposit' },
  withdraw: { label: 'Withdrawal', sign: '-', isPos: false, badgeClass: 'badge-withdraw' },
  transfer_in: { label: 'Transfer In', sign: '+', isPos: true, badgeClass: 'badge-transfer-in' },
  transfer_out: { label: 'Transfer Out', sign: '-', isPos: false, badgeClass: 'badge-transfer-out' },
  bill_payment: { label: 'Bill Payment', sign: '-', isPos: false, badgeClass: 'badge-bill' },
  donation: { label: 'Donation', sign: '-', isPos: false, badgeClass: 'badge-donation' },
};

export default function TransactionsTable({ transactions }) {
  const [filter, setFilter] = useState('all');

  const filtered = transactions.filter((tx) => {
    if (filter === 'all') return true;
    if (filter === 'transfers') return tx.type === 'transfer_in' || tx.type === 'transfer_out';
    if (filter === 'money') return tx.type === 'deposit' || tx.type === 'withdraw';
    if (filter === 'bills_donations') return tx.type === 'bill_payment' || tx.type === 'donation';
    return true;
  });

  return (
    <section className="card">
      <div className="tx-header-row">
        <h2>Recent transactions</h2>

        <div className="filter-pills">
          <button
            type="button"
            className={`pill ${filter === 'all' ? 'active' : ''}`}
            onClick={() => setFilter('all')}
          >
            All ({transactions.length})
          </button>
          <button
            type="button"
            className={`pill ${filter === 'transfers' ? 'active' : ''}`}
            onClick={() => setFilter('transfers')}
          >
            Transfers
          </button>
          <button
            type="button"
            className={`pill ${filter === 'bills_donations' ? 'active' : ''}`}
            onClick={() => setFilter('bills_donations')}
          >
            Bills &amp; Donations
          </button>
          <button
            type="button"
            className={`pill ${filter === 'money' ? 'active' : ''}`}
            onClick={() => setFilter('money')}
          >
            Deposits &amp; Cash
          </button>
        </div>
      </div>

      <div className="table-wrap">
        <table className="tx-table">
          <thead>
            <tr>
              <th>Date</th>
              <th>Type</th>
              <th>Description</th>
              <th className="num">Amount</th>
              <th className="num">Balance</th>
            </tr>
          </thead>
          <tbody id="tx-body">
            {filtered.length === 0 ? (
              <tr>
                <td colSpan={5} className="muted center">
                  {transactions.length === 0
                    ? 'No transactions yet.'
                    : 'No transactions match this filter.'}
                </td>
              </tr>
            ) : (
              filtered.map((tx) => {
                const conf = TYPE_CONFIG[tx.type] || {
                  label: tx.type,
                  sign: '-',
                  isPos: false,
                  badgeClass: '',
                };

                return (
                  <tr key={tx.id}>
                    <td className="tx-date">{formatDate(tx.created_at)}</td>
                    <td>
                      <span className={`tx-type-badge ${conf.badgeClass}`}>
                        {conf.label}
                      </span>
                    </td>
                    <td className="tx-desc">
                      {tx.description || (conf.isPos ? 'Direct deposit' : 'Cash withdrawal')}
                    </td>
                    <td className={`num ${conf.isPos ? 'pos' : 'neg'}`}>
                      {conf.sign}
                      {formatMoney(tx.amount_cents)}
                    </td>
                    <td className="num">{formatMoney(tx.balance_cents)}</td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>
    </section>
  );
}
