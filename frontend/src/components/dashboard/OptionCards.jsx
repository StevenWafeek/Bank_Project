const icons = {
  balance: (
    <>
      <rect x="2" y="6" width="20" height="13" rx="2" />
      <path d="M2 10h20" />
      <circle cx="16" cy="15" r="1.5" />
    </>
  ),
  transfer: (
    <>
      <path d="M17 3L21 7L17 11" />
      <path d="M3 7H21" />
      <path d="M7 21L3 17L7 13" />
      <path d="M21 17H3" />
    </>
  ),
  deposit: (
    <>
      <path d="M12 5v14" />
      <path d="M19 12l-7 7-7-7" />
    </>
  ),
  withdraw: (
    <>
      <path d="M12 19V5" />
      <path d="M5 12l7-7 7 7" />
    </>
  ),
  bills: (
    <>
      <path d="M14 2H6a2 2 0 0 0-2 2v16l3-2 3 2 3-2 3 2 3-2 3 2V8z" />
      <path d="M14 2v6h6" />
      <path d="M8 12h8" />
      <path d="M8 16h5" />
    </>
  ),
  donate: (
    <>
      <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z" />
    </>
  ),
  manage: (
    <>
      <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
      <circle cx="12" cy="7" r="4" />
    </>
  ),
};

const options = [
  { key: 'balance', label: 'Your Balance' },
  { key: 'transfer', label: 'Transfer' },
  { key: 'deposit', label: 'Deposit' },
  { key: 'withdraw', label: 'Withdraw' },
  { key: 'bills', label: 'Pay Bills' },
  { key: 'donate', label: 'Donations' },
  { key: 'manage', label: 'Manage Account' },
];

export default function OptionCards({ active, onSelect }) {
  return (
    <section className="options">
      {options.map(({ key, label }) => (
        <button
          key={key}
          className={`option ${active === key ? 'active' : ''}`}
          data-panel={key}
          onClick={() => onSelect(key)}
          type="button"
        >
          <svg
            className="option-icon"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
            aria-hidden="true"
          >
            {icons[key]}
          </svg>
          <span>{label}</span>
        </button>
      ))}
    </section>
  );
}
