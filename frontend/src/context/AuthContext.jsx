import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import { api } from '../api/client.js';

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  const refreshUser = useCallback(async () => {
    try {
      const u = await api.me();
      setUser(u);
      return u;
    } catch {
      setUser(null);
      return null;
    }
  }, []);

  useEffect(() => {
    refreshUser().finally(() => setLoading(false));
  }, [refreshUser]);

  const login = useCallback(async (email, password) => {
    const u = await api.login(email, password);
    setUser(u);
    return u;
  }, []);

  const signup = useCallback(async (name, email, password) => {
    const u = await api.signup(name, email, password);
    setUser(u);
    return u;
  }, []);

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
    () => ({ user, setUser, refreshUser, loading, login, signup, logout, handleAuthError }),
    [user, refreshUser, loading, login, signup, logout, handleAuthError],
  );
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export const useAuth = () => useContext(AuthContext);
