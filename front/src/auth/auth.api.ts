import { http } from "@/lib/api/http";
import type { AuthResponse, LoginPayload, RegisterPayload, User } from "./auth.types";

export const authApi = {
  login(payload: LoginPayload) {
    return http<AuthResponse>("/auth/login", { method: "POST", json: payload });
  },

  register(payload: RegisterPayload) {
    return http<AuthResponse>("/auth/register", { method: "POST", json: payload });
  },

  me() {
    return http<{ user: User }>("/auth/me", { method: "GET" });
  },

  logout() {
    return http<void>("/auth/logout", { method: "POST" });
  },

  // Pour OAuth Discord : en SPA, tu rediriges vers le backend
  // Exemple: window.location.href = `${API_URL}/auth/discord`
  // Ici on expose juste l'URL pratique.
  discordStartUrl() {
    const API_URL = import.meta.env.VITE_API_URL ?? "";
    return `${API_URL}/auth/discord`;
  },
};