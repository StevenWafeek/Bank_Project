import { useState } from 'react';

export default function SignupForm({ onSubmit, busy }) {
  const [form, setForm] = useState({ name: '', email: '', password: '', confirm: '' });
  const update = (e) => setForm({ ...form, [e.target.name]: e.target.value });

  function handleSubmit(e) {
    e.preventDefault();
    onSubmit(() => {
      if (!form.name || !form.email || !form.password) throw new Error('Please fill in all fields.');
      if (form.password.length < 8) throw new Error('Password must be at least 8 characters.');
      if (form.password !== form.confirm) throw new Error('Passwords do not match.');
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
      <button type="submit" className="btn btn-primary" disabled={busy}>
        Create account
      </button>
    </form>
  );
}
