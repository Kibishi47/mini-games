import { createContext, useContext } from "react";
import type { User } from "./auth.types";

export type AuthState = {
    user: User | null;
    accessToken: string | null;
    isLoading: boolean;
}

export type AuthActions = {
    setAuth: (data: { user: User; accessToken: string }) => void;
    clearAuth: () => void;
    refreshAuth: () => Promise<void>;
    logout: () => Promise<void>;
}

export type AuthContextType = AuthState & AuthActions;

export const AuthContext = createContext<AuthContextType | null>(null);

export const useAuth = () => {
    const context = useContext(AuthContext);
    if (!context) {
        throw new Error("useAuth must be used within an AuthProvider");
    }
    return context;
};