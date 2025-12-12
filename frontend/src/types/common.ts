/**
 * Common type definitions shared across the application
 */

// Base types
export interface TimestampedEntity {
  created_at: string;
  updated_at: string;
}

export interface EntityWithId {
  id: number;
}

export interface BaseEntity extends EntityWithId, TimestampedEntity {}

// API response types
export interface ApiResponse<T> {
  data: T;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  per_page: number;
}

// Error types
export interface ValidationErrors {
  [key: string]: string[];
}

export interface ApiErrorResponse {
  error: string;
  message?: string;
  errors?: ValidationErrors;
}

// Status response format used by backend
export interface StatusResponse<T = unknown> {
  status: {
    code: number;
    message: string;
  };
  data?: T;
}

// Common data operation types
export type CreateData<T> = Omit<T, keyof BaseEntity>;
export type UpdateData<T> = Partial<Omit<T, keyof EntityWithId>>;

// Filter and sort types
export type SortDirection = "asc" | "desc";
export interface SortOptions<T> {
  field: keyof T;
  direction: SortDirection;
}
