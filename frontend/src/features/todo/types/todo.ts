import type { BaseEntity, ValidationErrors } from "@/types/common";
import type { Category } from "@/features/category/types/category";
import type { Tag } from "@/features/tag/types/tag";

/**
 * Todo domain types
 */

// Enums
export type TodoPriority = "low" | "medium" | "high";
export type TodoStatus = "pending" | "in_progress" | "completed";
export type TodoFilter = "all" | "active" | "completed";

// Category reference (simplified version for todos)
export type TodoCategoryRef = Pick<Category, "id" | "name" | "color">;

// Tag reference (simplified version for todos)
export type TodoTagRef = Pick<Tag, "id" | "name" | "color">;

// File type (Active Storage format)
export interface TodoFile {
  id: string | number;
  filename: string;
  content_type: string;
  byte_size: number;
  url: string;
  variants?: {
    thumb?: string;
    medium?: string;
  };
}

// Main todo entity
export interface Todo extends BaseEntity {
  title: string;
  completed: boolean;
  position: number;
  due_date: string | null;
  priority: TodoPriority;
  status: TodoStatus;
  description: string | null;
  category: TodoCategoryRef | null;
  tags: TodoTagRef[];
  files: TodoFile[];
  comments_count: number;
  latest_comments: unknown[]; // 最新のコメント（詳細表示用）
  history_count: number;
  highlights?: {
    title?: Array<{ start: number; end: number; matched_text: string }>;
    description?: Array<{ start: number; end: number; matched_text: string }>;
  };
}

// Todo operations
export interface CreateTodoData {
  title: string;
  due_date?: string | null;
  priority?: TodoPriority;
  status?: TodoStatus;
  description?: string | null;
  category_id?: number | null;
  tag_ids?: number[];
}

export interface UpdateTodoData {
  title?: string;
  completed?: boolean;
  due_date?: string | null;
  priority?: TodoPriority;
  status?: TodoStatus;
  description?: string | null;
  category_id?: number | null;
  tag_ids?: number[];
}

export interface UpdateOrderData {
  id: number;
  position: number;
}

// Error types
export type TodoValidationErrors = ValidationErrors;

// For backward compatibility (can be removed after migration)
export interface TodoCategory {
  id: number;
  name: string;
  color: string;
}

export interface TodoError {
  title?: string[];
  priority?: string[];
  status?: string[];
  description?: string[];
  due_date?: string[];
  base?: string[];
}

// Search and filter types
export interface TodoSearchParams {
  q?: string; // Search query
  category_id?: number | number[] | null;
  status?: TodoStatus | TodoStatus[];
  priority?: TodoPriority | TodoPriority[];
  tag_ids?: number[];
  tag_mode?: "any" | "all"; // How to match tags
  due_date_from?: string;
  due_date_to?: string;
  sort_by?: "position" | "created_at" | "updated_at" | "due_date" | "title" | "priority" | "status";
  sort_order?: "asc" | "desc";
  page?: number;
  per_page?: number;
}

export interface TodoSearchResponse {
  data: Todo[];
  meta: {
    total: number;
    current_page: number;
    total_pages: number;
    per_page: number;
    search_query?: string;
    filters_applied: Record<string, unknown>;
  };
  suggestions?: Array<{
    type: string;
    message: string;
    current_filters?: string[];
  }>;
}

export interface ActiveFilters {
  search?: string;
  category_id?: number | null;
  status?: TodoStatus[];
  priority?: TodoPriority[];
  tag_ids?: number[];
  date_range?: {
    from?: string;
    to?: string;
  };
}
