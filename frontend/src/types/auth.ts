import type { BaseEntity } from "./common";

/**
 * Authentication and user related types
 */

export interface User extends BaseEntity {
  email: string;
  name: string;
}

// API request types
export interface LoginRequest {
  user: {
    email: string;
    password: string;
  };
}

export interface RegisterRequest {
  user: {
    email: string;
    password: string;
    password_confirmation: string;
    name: string;
  };
}

// API response types
export interface AuthResponse {
  status: {
    code: number;
    message: string;
  };
  data: User;
}

export interface AuthError {
  status: {
    code: number;
    message: string;
  };
}

// Authentication result type
export interface AuthResult {
  user: User;
  token: string;
}
