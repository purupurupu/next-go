import { httpClient } from "./api-client";
import { API_BASE_URL, API_ENDPOINTS } from "./constants";
import type {
  User,
  LoginRequest,
  RegisterRequest,
  AuthResponse,
  AuthResult,
} from "@/types/auth";

// Re-export for backward compatibility
export type { User, LoginRequest, RegisterRequest, AuthResponse, AuthResult };

class AuthClient {
  private async authRequest(endpoint: string, data: unknown): Promise<AuthResult> {
    try {
      const response = await fetch(`${API_BASE_URL}${endpoint}`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify(data),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || errorData.status?.message || "Authentication failed");
      }

      const responseData: AuthResponse = await response.json();

      // Authorizationヘッダーからトークンを取得
      const authHeader = response.headers.get("Authorization");
      if (!authHeader) {
        throw new Error("No authorization token received");
      }

      // Bearerプレフィックスを削除
      const token = authHeader.replace("Bearer ", "");

      if (token) {
        this.setAuthToken(token);
        // ユーザー情報も保存
        this.setUser(responseData.data);
      }

      return {
        user: responseData.data,
        token: token,
      };
    } catch (error) {
      if (error instanceof Error) {
        throw error;
      }
      throw new Error("Network error occurred");
    }
  }

  async login(credentials: LoginRequest): Promise<AuthResult> {
    return this.authRequest(API_ENDPOINTS.AUTH_LOGIN, credentials);
  }

  async register(userData: RegisterRequest): Promise<AuthResult> {
    return this.authRequest(API_ENDPOINTS.AUTH_REGISTER, userData);
  }

  async logout(): Promise<void> {
    try {
      await httpClient.delete(API_ENDPOINTS.AUTH_LOGOUT);
    } finally {
      this.removeAuthToken();
    }
  }

  setAuthToken(token: string): void {
    if (typeof window !== "undefined") {
      // Remove 'Bearer ' prefix if it exists
      const cleanToken = token.replace("Bearer ", "");
      localStorage.setItem("authToken", cleanToken);
    }
  }

  getAuthToken(): string | null {
    if (typeof window === "undefined") return null;
    return localStorage.getItem("authToken");
  }

  removeAuthToken(): void {
    if (typeof window !== "undefined") {
      localStorage.removeItem("authToken");
      localStorage.removeItem("user");
    }
  }

  setUser(user: User): void {
    if (typeof window !== "undefined") {
      localStorage.setItem("user", JSON.stringify(user));
    }
  }

  getUser(): User | null {
    if (typeof window === "undefined") return null;
    const userStr = localStorage.getItem("user");
    if (userStr) {
      try {
        return JSON.parse(userStr);
      } catch {
        return null;
      }
    }
    return null;
  }

  isAuthenticated(): boolean {
    return !!this.getAuthToken();
  }
}

export const authClient = new AuthClient();
