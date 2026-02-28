import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  type ReactNode,
} from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { createElement } from "react";
import i18next from "i18next";
import axios from "axios";
import { getMe, logout as apiLogout } from "@/api/auth";
import type { User } from "@/types";

interface AuthContextValue {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue>({
  user: null,
  isLoading: true,
  isAuthenticated: false,
  logout: async () => {},
});

export function AuthProvider({ children }: { children: ReactNode }) {
  const qc = useQueryClient();
  const { data: user, isLoading } = useQuery({
    queryKey: ["auth", "me"],
    queryFn: getMe,
    retry: false,
    staleTime: 5 * 60 * 1000,
  });

  // Sync language preference on login
  useEffect(() => {
    if (!user) return;
    const localLang = i18next.language?.startsWith("pl") ? "pl" : "en";
    const profileLang = user.preferred_language || "en";

    if (profileLang !== "en" && profileLang !== localLang) {
      // Server has a non-default preference — adopt it
      i18next.changeLanguage(profileLang);
    } else if (profileLang === "en" && localLang !== "en") {
      // User chose a language locally but profile is default — sync to server
      axios.patch("/api/v1/auth/me", { preferred_language: localLang }).catch(() => {});
    }
  }, [user]);

  const logout = useCallback(async () => {
    await apiLogout();
    qc.setQueryData(["auth", "me"], null);
    qc.clear();
    globalThis.location.assign("/login");
  }, [qc]);

  return createElement(
    AuthContext.Provider,
    {
      value: {
        user: user ?? null,
        isLoading,
        isAuthenticated: !!user,
        logout,
      },
    },
    children,
  );
}

export function useAuth() {
  return useContext(AuthContext);
}
