"use client";

import React, { createContext, useContext, useEffect, useState } from "react";
import { authClient, User, LoginRequest, RegisterRequest } from "@/lib/auth-client";

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (credentials: LoginRequest) => Promise<void>;
  register: (userData: RegisterRequest) => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [hasToken, setHasToken] = useState(false);

  useEffect(() => {
    // Check if user is already logged in
    const token = authClient.getAuthToken();
    const savedUser = authClient.getUser();

    if (token) {
      setHasToken(true);
      if (savedUser) {
        setUser(savedUser);
      }
    }
    setIsLoading(false);
  }, []);

  const login = async (credentials: LoginRequest) => {
    setIsLoading(true);
    try {
      const { user: loggedInUser } = await authClient.login(credentials);
      setUser(loggedInUser);
      setHasToken(true);
    } catch (error) {
      throw error;
    } finally {
      setIsLoading(false);
    }
  };

  const register = async (userData: RegisterRequest) => {
    setIsLoading(true);
    try {
      const { user: registeredUser } = await authClient.register(userData);
      setUser(registeredUser);
      setHasToken(true);
    } catch (error) {
      throw error;
    } finally {
      setIsLoading(false);
    }
  };

  const logout = async () => {
    setIsLoading(true);
    try {
      await authClient.logout();
    } catch {
      // Silently handle logout errors - we'll clear the local state anyway
    } finally {
      // Always clear local state regardless of server response
      setUser(null);
      setHasToken(false);
      setIsLoading(false);
    }
  };

  const isAuthenticated = !!user || hasToken;

  const value: AuthContextType = {
    user,
    isAuthenticated,
    isLoading,
    login,
    register,
    logout,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextType {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
