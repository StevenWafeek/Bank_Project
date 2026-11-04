import { Link, NavLink, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext.jsx';

export default function Navbar() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  async function handleLogout() {
    await logout();
    navigate('/', { replace: true });
  }

  return (
    <header className="navbar">
      <Link to={user ? '/dashboard' : '/'} className="brand">
        <span className="logo">M</span>MIeBank
      </Link>
      <nav>
        <NavLink id="home-link" to={user ? '/dashboard' : '/'} end>
          {user ? 'Dashboard' : 'Log in'}
        </NavLink>
        <NavLink to="/about">About</NavLink>
        {user && (
          <button id="logout-btn" className="btn btn-ghost" onClick={handleLogout}>
            Log out
          </button>
        )}
      </nav>
    </header>
  );
}
