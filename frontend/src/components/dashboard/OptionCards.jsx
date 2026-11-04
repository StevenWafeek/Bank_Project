const icons = {
  balance: (
    <>
      <rect x="2" y="6" width="20" height="13" rx="2" />
      <path d="M2 10h20" />
      <circle cx="16" cy="15" r="1.5" />
    </>
  ),
  withdraw: (
    <>
      <path d="M12 19V5" />
      <path d="M5 12l7-7 7 7" />
    </>
  ),
  deposit: (
    <>
      <path d="M12 5v14" />
      <path d="M19 12l-7 7-7-7" />
    </>
  ),
};

const options = [
  { key: 'balance', label: 'Your Balance' },
  { key: 'withdraw', label: 'Withdraw' },
  { key: 'deposit', label: 'Deposit' },
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
