const values = [
  {
    title: 'Our Mission',
    text: 'To give everyone fast, fee-free access to their money with tools that are easy to understand and a pleasure to use.',
  },
  {
    title: 'Security First',
    text: 'Passwords are hashed with bcrypt, sessions use secure HTTP-only cookies, and every transaction is validated on our servers.',
  },
  {
    title: 'Always Available',
    text: 'Check your balance, deposit or withdraw funds 24/7 from any device — no branch visits, no waiting in line.',
  },
];

const stats = [
  ['2026', 'Founded'],
  ['24/7', 'Online banking'],
  ['$0', 'Monthly fees'],
  ['100%', 'Digital'],
];

export default function AboutPage() {
  return (
    <main className="container about">
      <section className="about-hero">
        <h1>
          About <span className="accent">MIeBank</span>
        </h1>
        <p>
          MIeBank is a modern digital bank built around one idea: managing your money should be simple, transparent
          and secure.
        </p>
      </section>

      <section className="grid-3">
        {values.map((v) => (
          <article className="card" key={v.title}>
            <h3>{v.title}</h3>
            <p>{v.text}</p>
          </article>
        ))}
      </section>

      <section className="card stats">
        {stats.map(([value, label]) => (
          <div key={label}>
            <strong>{value}</strong>
            <span>{label}</span>
          </div>
        ))}
      </section>

      <section className="card">
        <h3>Contact us</h3>
        <p>
          Email: <a href="mailto:support@miebank.com">support@miebank.com</a>
          <br />
          Phone: +1 (800) 643-2265
          <br />
          Address: 1 MIeBank Plaza, Financial District
        </p>
      </section>
    </main>
  );
}
