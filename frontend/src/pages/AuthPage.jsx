import { useState } from 'react';
import { Navigate, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext.jsx';
import LoginForm from '../components/auth/LoginForm.jsx';
import SignupForm from '../components/auth/SignupForm.jsx';
import Message from '../components/Message.jsx';

export default function AuthPage() {
  const { user, loading, login, signup } = useAuth();
  const navigate = useNavigate();
  const [tab, setTab] = useState('login');
  const [message, setMessage] = useState('');
  const [busy, setBusy] = useState(false);

  if (loading) return null;
  if (user) return <Navigate to="/dashboard" replace />;

  function switchTab(next) {
    setTab(next);
    setMessage('');
  }

  // validate() returns the form values or throws a user-facing error.
  async function submit(validate) {
    setMessage('');
    setBusy(true);
    try {
      const f = validate();
      if (tab === 'login') await login(f.email, f.password);
      else await signup(f.name, f.email, f.password);
      navigate('/dashboard', { replace: true });
    } catch (err) {
      setMessage(err.message);
    } finally {
      setBusy(false);
    }
  }

  return (
    <main className="auth-wrapper">
      <section className="auth-hero">
        <h1>
          Banking made <span className="accent">simple</span>.
        </h1>
        <p>Check your balance, deposit and withdraw money securely — anytime, anywhere with MIeBank.</p>
        <ul className="features">
          <li>Secure, encrypted passwords</li>
          <li>Instant deposits &amp; withdrawals</li>
          <li>Full transaction history</li>
        </ul>
      </section>

      <section className="card auth-card">
        <div className="tabs" role="tablist">
          {['login', 'signup'].map((t) => (
            <button
              key={t}
              className={`tab ${tab === t ? 'active' : ''}`}
              data-tab={t}
              role="tab"
              onClick={() => switchTab(t)}
            >
              {t === 'login' ? 'Log in' : 'Sign up'}
            </button>
          ))}
        </div>
        {tab === 'login' ? (
          <LoginForm onSubmit={submit} busy={busy} />
        ) : (
          <SignupForm onSubmit={submit} busy={busy} />
        )}
        <Message id="auth-message" text={message} />
      </section>
    </main>
  );
}
