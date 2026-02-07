import { useCallback, useEffect, useMemo, useState } from "react";
import type { User } from "./auth.types";
import { authApi } from "./auth.api";
import { AuthContext } from "./AuthContext";

export function AuthProvider({children}: {children: React.ReactNode}) {
    const [user, setUser] = useState<User | null>(null);
    const [accessToken, setAccessToken] = useState<string | null>(null);
    const [isLoading, setIsLoading] = useState(true);

    const setAuth = (data: { user: User; accessToken: string }) => {
        setUser(data.user);
        setAccessToken(data.accessToken);
    };

    const clearAuth = () => {
        setUser(null);
        setAccessToken(null);
    };

    const refreshAuth = useCallback(async () => {
        const data = await authApi.refresh();
        setAuth({ user: data.user, accessToken: data.accessToken });
    }, [setAuth]);

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
        clearAuth();
      } finally {
        setIsLoading(false);
      }
    })();
  }, [refreshAuth, clearAuth]);

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