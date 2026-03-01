import { useCallback, useEffect, useMemo, useState } from "react";
import type { User } from "./auth.types";
import { authApi } from "./auth.api";
import { AuthContext } from "./AuthContext";
import { HttpError } from "@/lib/api/http";

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [accessToken, setAccessToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const clearAuth = useCallback(() => {
    setUser(null);
    setAccessToken(null);
    localStorage.removeItem("accessToken");
  }, []);

  const refreshAuth = useCallback(async () => {
    const token = localStorage.getItem("accessToken");
    if (!token) {
      clearAuth();
      return;
    }

    try {
      const { user } = await authApi.me();
      setUser(user);
      setAccessToken(token);
    } catch (error) {
      if (error instanceof HttpError && (error.status === 401 || error.status === 403)) {
        clearAuth();
      }
      throw error;
    }
  }, [clearAuth]);

  const setAuth = useCallback(async (token: string) => {
    localStorage.setItem("accessToken", token);
    setAccessToken(token);
    try {
      const { user } = await authApi.me();
      setUser(user);
    } catch (error) {
      if (error instanceof HttpError && (error.status === 401 || error.status === 403)) {
        clearAuth();
      }
      throw error;
    }
  }, [clearAuth]);

  const logout = useCallback(async () => {
    try {
      await authApi.logout();
    } finally {
      clearAuth();
    }
  }, [clearAuth]);

  useEffect(() => {
    (async () => {
      try {
        await refreshAuth();
      } catch {
        // User not logged in or token expired
      } finally {
        setIsLoading(false);
      }
    })();
  }, [refreshAuth]);

  const value = useMemo(() => {
    return {
      user,
      accessToken,
      isLoading,
      setAuth,
      clearAuth,
      refreshAuth,
      logout,
    }
  }, [user, accessToken, isLoading, setAuth, clearAuth, refreshAuth, logout]);

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}