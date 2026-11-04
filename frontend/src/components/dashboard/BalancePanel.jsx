import { formatMoney } from '../../utils/format.js';

export default function BalancePanel({ balanceCents }) {
  return (
    <section className="card panel" id="panel-balance">
      <h2>Your Balance</h2>
      <p className="balance" id="balance-amount">
        {formatMoney(balanceCents)}
      </p>
      <p className="muted">Available balance</p>
    </section>
  );
}
