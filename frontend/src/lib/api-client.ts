import { API_BASE_URL } from "./constants";

class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public errors?: Record<string, string[]>,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

class HttpClient {
  private baseUrl: string;

  constructor(baseUrl: string = API_BASE_URL) {
    this.baseUrl = baseUrl;
  }

  private getAuthToken(): string | null {
    if (typeof window === "undefined") return null;
    return localStorage.getItem("authToken");
  }

  private getAuthHeaders(): Record<string, string> {
    const token = this.getAuthToken();
    if (token) {
      return { Authorization: `Bearer ${token}` };
    }
    return {};
  }

  private async request<T>(
    endpoint: string,
    options?: RequestInit,
  ): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`;

    const config: RequestInit = {
      headers: {
        "Content-Type": "application/json",
        ...this.getAuthHeaders(),
        ...options?.headers,
      },
      credentials: "include",
      ...options,
    };

    try {
      const response = await fetch(url, config);

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));

        // Handle v1 API error format
        if (errorData && typeof errorData === "object" && "error" in errorData) {
          const error = errorData.error;
          const message = error.message || `HTTP ${response.status}: ${response.statusText}`;
          throw new ApiError(
            message,
            response.status,
            error.details || error,
          );
        }

        throw new ApiError(
          `HTTP ${response.status}: ${response.statusText}`,
          response.status,
          errorData,
        );
      }

      // Handle no content responses
      if (response.status === 204) {
        return null as T;
      }

      return await response.json() as T;
    } catch (error) {
      if (error instanceof ApiError) {
        throw error;
      }
      throw new ApiError("Network error occurred", 0);
    }
  }

  async get<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint);
  }

  async post<T>(endpoint: string, data?: unknown): Promise<T> {
    return this.request<T>(endpoint, {
      method: "POST",
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  async put<T>(endpoint: string, data?: unknown): Promise<T> {
    return this.request<T>(endpoint, {
      method: "PUT",
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  async patch<T>(endpoint: string, data?: unknown): Promise<T> {
    return this.request<T>(endpoint, {
      method: "PATCH",
      body: data ? JSON.stringify(data) : undefined,
    });
  }

  async delete<T>(endpoint: string): Promise<T> {
    return this.request<T>(endpoint, {
      method: "DELETE",
    });
  }

  async uploadFile<T>(endpoint: string, formData: FormData, method: string = "POST"): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`;

    const config: RequestInit = {
      method,
      headers: {
        ...this.getAuthHeaders(),
        // Don't set Content-Type - let browser set it with boundary for multipart/form-data
      },
      credentials: "include",
      body: formData,
    };

    try {
      const response = await fetch(url, config);

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));

        // Handle v1 API error format
        if (errorData && typeof errorData === "object" && "error" in errorData) {
          const error = errorData.error;
          const message = error.message || `HTTP ${response.status}: ${response.statusText}`;
          throw new ApiError(
            message,
            response.status,
            error.details || error,
          );
        }

        throw new ApiError(
          `HTTP ${response.status}: ${response.statusText}`,
          response.status,
          errorData,
        );
      }

      return await response.json() as T;
    } catch (error) {
      if (error instanceof ApiError) {
        throw error;
      }
      throw new ApiError("Network error occurred", 0);
    }
  }
}

// API v1用のベースクライアント（/api/v1プレフィックス付き）
class ApiClient extends HttpClient {
  constructor() {
    super(`${API_BASE_URL}/api/v1`);
  }
}

export const httpClient = new HttpClient();
export const apiClient = new ApiClient();
export { ApiError, HttpClient, ApiClient };
export type { HttpClient as HttpClientType };
