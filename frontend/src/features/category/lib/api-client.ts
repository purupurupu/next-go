import { HttpClient, ApiError } from "@/lib/api-client";
import { API_ENDPOINTS } from "@/lib/constants";
import type { Category, CreateCategoryData, UpdateCategoryData } from "../types/category";

class CategoryApiClient extends HttpClient {
  async getCategories(): Promise<Category[]> {
    const response = await this.get<Category[]>(API_ENDPOINTS.CATEGORIES);
    // 配列であることを保証
    return Array.isArray(response) ? response : [];
  }

  async getCategory(id: number): Promise<Category> {
    return this.get<Category>(API_ENDPOINTS.CATEGORY_BY_ID(id));
  }

  async createCategory(data: CreateCategoryData): Promise<Category> {
    return this.post<Category>(API_ENDPOINTS.CATEGORIES, {
      category: data,
    });
  }

  async updateCategory(id: number, data: UpdateCategoryData): Promise<Category> {
    return this.patch<Category>(API_ENDPOINTS.CATEGORY_BY_ID(id), {
      category: data,
    });
  }

  async deleteCategory(id: number): Promise<void> {
    return this.delete<void>(API_ENDPOINTS.CATEGORY_BY_ID(id));
  }
}

export const categoryApiClient = new CategoryApiClient();
export { ApiError };
