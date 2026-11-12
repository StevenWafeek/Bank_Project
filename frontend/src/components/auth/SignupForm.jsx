import { useState } from 'react';

export default function SignupForm({ onSubmit, busy }) {
  const [form, setForm] = useState({
    name: '',
    email: '',
    password: '',
    confirm: '',
    card_number: '',
    card_expiry: '',
    iban: '',
  });
  const update = (e) => setForm({ ...form, [e.target.name]: e.target.value });

  function handleSubmit(e) {
    e.preventDefault();
    onSubmit(() => {
      if (!form.name || !form.email || !form.password || !form.card_number || !form.card_expiry || !form.iban) {
        throw new Error('Please fill in all fields.');
      }
      if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email)) {
        throw new Error('Please enter a valid email address, such as name@example.com.');
      }
      if (form.password.length < 8) throw new Error('Password must be at least 8 characters.');
      if (form.password !== form.confirm) throw new Error('Passwords do not match.');
      if (form.card_number.replace(/\D/g, '').length !== 16) {
        throw new Error('Card number must contain exactly 16 digits.');
      }
      if (!/^(0[1-9]|1[0-2])\/\d{2}$/.test(form.card_expiry)) {
        throw new Error('Card expiry must be in MM/YY format.');
      }
      return form;
    });
  }

  return (
    <form id="signup-form" className="form" noValidate onSubmit={handleSubmit}>
      <h2>Open your account</h2>
      <label>
        Full name
        <input type="text" name="name" autoComplete="name" maxLength={60} value={form.name} onChange={update} required />
      </label>
      <label>
        Email
        <input type="email" name="email" autoComplete="email" value={form.email} onChange={update} required />
      </label>
      <label>
        Password <small>(min. 8 characters)</small>
        <input
          type="password"
          name="password"
          autoComplete="new-password"
          maxLength={72}
          value={form.password}
          onChange={update}
          required
        />
      </label>
      <label>
        Confirm password
        <input
          type="password"
          name="confirm"
          autoComplete="new-password"
          value={form.confirm}
          onChange={update}
          required
        />
      </label>
      <label>
        Card number
        <input
          type="text"
          name="card_number"
          inputMode="numeric"
          autoComplete="cc-number"
          placeholder="16-digit card number"
          maxLength={23}
          value={form.card_number}
          onChange={update}
          required
        />
      </label>
      <label>
        Card expiry <small>(MM/YY)</small>
        <input
          type="text"
          name="card_expiry"
          inputMode="numeric"
          autoComplete="cc-exp"
          placeholder="MM/YY"
          maxLength={5}
          value={form.card_expiry}
          onChange={update}
          required
        />
      </label>
      <label>
        IBAN
        <input
          type="text"
          name="iban"
          autoComplete="off"
          placeholder="e.g. EG96MIEB0000000012345678"
          maxLength={34}
          value={form.iban}
          onChange={update}
          required
        />
      </label>
      <button type="submit" className="btn btn-primary" disabled={busy}>
        Create account
      </button>
    </form>
  );
}
