import { formatDate, formatMoney } from '../../utils/format.js';

export default function TransactionsTable({ transactions }) {
  return (
    <section className="card">
      <h2>Recent transactions</h2>
      <div className="table-wrap">
        <table className="tx-table">
          <thead>
            <tr>
              <th>Date</th>
              <th>Type</th>
              <th className="num">Amount</th>
              <th className="num">Balance</th>
            </tr>
          </thead>
          <tbody id="tx-body">
            {transactions.length === 0 ? (
              <tr>
                <td colSpan={4} className="muted center">
                  No transactions yet.
                </td>
              </tr>
            ) : (
              transactions.map((tx) => {
                const isDeposit = tx.type === 'deposit';
                return (
                  <tr key={tx.id}>
                    <td>{formatDate(tx.created_at)}</td>
                    <td>{isDeposit ? 'Deposit' : 'Withdraw'}</td>
                    <td className={`num ${isDeposit ? 'pos' : 'neg'}`}>
                      {isDeposit ? '+' : '-'}
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
