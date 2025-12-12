import type { BaseEntity, CreateData, UpdateData, ValidationErrors } from "@/types/common";

/**
 * Category domain types
 */

export interface Category extends BaseEntity {
  name: string;
  color: string;
  todo_count: number;
}

// Category operations
export type CreateCategoryData = CreateData<Pick<Category, "name" | "color">>;
export type UpdateCategoryData = UpdateData<Pick<Category, "name" | "color">>;

// Error types
export type CategoryValidationErrors = ValidationErrors;

// For backward compatibility (can be removed after migration)
export interface CategoryError {
  name?: string[];
  color?: string[];
  base?: string[];
}
