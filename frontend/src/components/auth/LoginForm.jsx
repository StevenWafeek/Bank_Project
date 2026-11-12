import { useState } from 'react';

export default function LoginForm({ onSubmit, busy }) {
  const [form, setForm] = useState({ email: '', password: '' });
  const update = (e) => setForm({ ...form, [e.target.name]: e.target.value });

  function handleSubmit(e) {
    e.preventDefault();
    onSubmit(() => {
      if (!form.email || !form.password) throw new Error('Please enter your email or card number and password.');
      return form;
    });
  }

  return (
    <form id="login-form" className="form" noValidate onSubmit={handleSubmit}>
      <h2>Welcome back</h2>
      <label>
        Email or Card Number
        <input
          type="text"
          name="email"
          autoComplete="username"
          placeholder="e.g. name@example.com or 16-digit card number"
          value={form.email}
          onChange={update}
          required
        />
      </label>
      <label>
        Password
        <input
          type="password"
          name="password"
          autoComplete="current-password"
          value={form.password}
          onChange={update}
          required
        />
      </label>
      <button type="submit" className="btn btn-primary" disabled={busy}>
        Log in
      </button>
    </form>
  );
}
