import { useState, useEffect, useCallback } from "react";
import { categoryApiClient, ApiError } from "../lib/api-client";
import type { Category, CreateCategoryData, UpdateCategoryData } from "../types/category";
import { toast } from "sonner";

export function useCategories(fetchOnMount = true) {
  const [categories, setCategories] = useState<Category[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const fetchCategories = useCallback(async () => {
    try {
      setIsLoading(true);
      setError(null);
      const data = await categoryApiClient.getCategories();
      setCategories(data);
    } catch (err) {
      const error = err instanceof ApiError ? err : new Error("Failed to fetch categories");
      setError(error);
      toast.error(error.message);
    } finally {
      setIsLoading(false);
    }
  }, []);

  const createCategory = useCallback(async (data: CreateCategoryData) => {
    try {
      const newCategory = await categoryApiClient.createCategory(data);
      setCategories((prev) => [...prev, newCategory]);
      toast.success("カテゴリーを作成しました");
      return newCategory;
    } catch (err) {
      const error = err instanceof ApiError ? err : new Error("Failed to create category");
      toast.error(error.message);
      throw error;
    }
  }, []);

  const updateCategory = useCallback(async (id: number, data: UpdateCategoryData) => {
    try {
      const updatedCategory = await categoryApiClient.updateCategory(id, data);
      setCategories((prev) =>
        prev.map((cat) => (cat.id === id ? updatedCategory : cat)),
      );
      toast.success("カテゴリーを更新しました");
      return updatedCategory;
    } catch (err) {
      const error = err instanceof ApiError ? err : new Error("Failed to update category");
      toast.error(error.message);
      throw error;
    }
  }, []);

  const deleteCategory = useCallback(async (id: number) => {
    try {
      await categoryApiClient.deleteCategory(id);
      setCategories((prev) => prev.filter((cat) => cat.id !== id));
      toast.success("カテゴリーを削除しました");
    } catch (err) {
      const error = err instanceof ApiError ? err : new Error("Failed to delete category");
      toast.error(error.message);
      throw error;
    }
  }, []);

  useEffect(() => {
    if (fetchOnMount) {
      fetchCategories();
    }
  }, [fetchOnMount]); // eslint-disable-line react-hooks/exhaustive-deps

  return {
    categories,
    isLoading,
    error,
    createCategory,
    updateCategory,
    deleteCategory,
    fetchCategories,
    refetch: fetchCategories,
  };
}
