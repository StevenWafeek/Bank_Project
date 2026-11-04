import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import { api } from '../api/client.js';

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.me().then(setUser).catch(() => setUser(null)).finally(() => setLoading(false));
  }, []);

  const login = useCallback(async (email, password) => setUser(await api.login(email, password)), []);
  const signup = useCallback(
    async (name, email, password) => setUser(await api.signup(name, email, password)),
    [],
  );
  const logout = useCallback(async () => {
    await api.logout().catch(() => {});
    setUser(null);
  }, []);

  // Drops the local user when the server says the session is gone; returns true if handled.
  const handleAuthError = useCallback((err) => {
    if (err?.status === 401) {
      setUser(null);
      return true;
    }
    return false;
  }, []);

  const value = useMemo(
    () => ({ user, loading, login, signup, logout, handleAuthError }),
    [user, loading, login, signup, logout, handleAuthError],
  );
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export const useAuth = () => useContext(AuthContext);
