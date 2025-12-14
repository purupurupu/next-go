import { HttpClient, ApiError } from "@/lib/api-client";
import type { Category, CreateCategoryData, UpdateCategoryData } from "../types/category";

class CategoryApiClient extends HttpClient {
  async getCategories(): Promise<Category[]> {
    const response = await this.get<Category[]>("/api/v1/categories");
    // 配列であることを保証
    return Array.isArray(response) ? response : [];
  }

  async getCategory(id: number): Promise<Category> {
    return this.get<Category>(`/api/v1/categories/${id}`);
  }

  async createCategory(data: CreateCategoryData): Promise<Category> {
    return this.post<Category>("/api/v1/categories", {
      category: data,
    });
  }

  async updateCategory(id: number, data: UpdateCategoryData): Promise<Category> {
    return this.patch<Category>(`/api/v1/categories/${id}`, {
      category: data,
    });
  }

  async deleteCategory(id: number): Promise<void> {
    return this.delete<void>(`/api/v1/categories/${id}`);
  }
}

export const categoryApiClient = new CategoryApiClient();
export { ApiError };
