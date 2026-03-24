"use client";

import React, { createContext, useContext, useState, useEffect, useCallback } from "react";
import { useRouter } from "next/navigation";
import { User } from "@/types";
import { apiFetch } from "./api";

interface AuthContextType {
  user: User | null;
  loading: boolean;
  login: (loginStr: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  isGuest: boolean;
  isAdmin: boolean;
  isManager: boolean;
}

const AuthContext = createContext<AuthContextType>(null!);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const router = useRouter();

  useEffect(() => {
    // Try to restore session from httpOnly cookie via /api/auth/me
    apiFetch<User>("/api/auth/me")
      .then(setUser)
      .catch(() => setUser(null))
      .finally(() => setLoading(false));
  }, []);

  const login = useCallback(async (loginStr: string, password: string) => {
    const res = await apiFetch<{ token: string; user: User }>("/api/auth/login", {
      method: "POST",
      body: JSON.stringify({ login: loginStr, password }),
    });
    setUser(res.user);
  }, []);

  const logout = useCallback(async () => {
    await apiFetch("/api/auth/logout", { method: "POST" }).catch(() => {});
    setUser(null);
    router.push("/login");
  }, [router]);

  const isGuest = !user;
  const isAdmin = user?.role === "admin";
  const isManager = user?.role === "manager";

  return (
    <AuthContext.Provider value={{ user, loading, login, logout, isGuest, isAdmin, isManager }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  return useContext(AuthContext);
}
