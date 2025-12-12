import { useState, useCallback, useMemo } from "react";
import type { TodoSearchParams, TodoStatus, TodoPriority, ActiveFilters } from "../types/todo";

interface UseSearchParamsReturn {
  searchParams: TodoSearchParams;
  activeFilters: ActiveFilters;
  hasActiveFilters: boolean;
  updateSearchQuery: (query: string) => void;
  updateCategory: (categoryId: number | null | undefined) => void;
  updateStatus: (status: TodoStatus[]) => void;
  updatePriority: (priority: TodoPriority[]) => void;
  updateTags: (tagIds: number[], tagMode?: "any" | "all") => void;
  updateDateRange: (from?: string, to?: string) => void;
  updateSort: (sortBy: TodoSearchParams["sort_by"], sortOrder?: TodoSearchParams["sort_order"]) => void;
  updatePage: (page: number) => void;
  clearFilters: () => void;
  clearSingleFilter: (filterType: keyof ActiveFilters) => void;
}

const defaultParams: TodoSearchParams = {
  sort_by: "position",
  sort_order: "asc",
  per_page: 50,
  page: 1,
};

export function useSearchParams(): UseSearchParamsReturn {
  const [searchParams, setSearchParams] = useState<TodoSearchParams>(() => ({ ...defaultParams }));

  // Update search query
  const updateSearchQuery = useCallback((query: string) => {
    setSearchParams((prev) => ({
      ...prev,
      q: query || undefined,
      page: 1, // Reset to first page on new search
    }));
  }, []);

  // Update category filter
  const updateCategory = useCallback((categoryId: number | null | undefined) => {
    setSearchParams((prev) => {
      const newParams = { ...prev };
      if (categoryId === undefined) {
        delete newParams.category_id;
      } else {
        newParams.category_id = categoryId;
      }
      newParams.page = 1;
      return newParams;
    });
  }, []);

  // Update status filter
  const updateStatus = useCallback((status: TodoStatus[]) => {
    setSearchParams((prev) => ({
      ...prev,
      status: status.length > 0 ? status : undefined,
      page: 1,
    }));
  }, []);

  // Update priority filter
  const updatePriority = useCallback((priority: TodoPriority[]) => {
    setSearchParams((prev) => ({
      ...prev,
      priority: priority.length > 0 ? priority : undefined,
      page: 1,
    }));
  }, []);

  // Update tag filter
  const updateTags = useCallback((tagIds: number[], tagMode: "any" | "all" = "any") => {
    setSearchParams((prev) => ({
      ...prev,
      tag_ids: tagIds.length > 0 ? tagIds : undefined,
      tag_mode: tagIds.length > 0 ? tagMode : undefined,
      page: 1,
    }));
  }, []);

  // Update date range filter
  const updateDateRange = useCallback((from?: string, to?: string) => {
    setSearchParams((prev) => ({
      ...prev,
      due_date_from: from || undefined,
      due_date_to: to || undefined,
      page: 1,
    }));
  }, []);

  // Update sort
  const updateSort = useCallback((sortBy: TodoSearchParams["sort_by"], sortOrder?: TodoSearchParams["sort_order"]) => {
    setSearchParams((prev) => ({
      ...prev,
      sort_by: sortBy,
      sort_order: sortOrder || prev.sort_order,
    }));
  }, []);

  // Update page
  const updatePage = useCallback((page: number) => {
    setSearchParams((prev) => ({
      ...prev,
      page,
    }));
  }, []);

  // Clear all filters
  const clearFilters = useCallback(() => {
    setSearchParams(defaultParams);
  }, []);

  // Clear single filter
  const clearSingleFilter = useCallback((filterType: keyof ActiveFilters) => {
    setSearchParams((prev) => {
      const newParams = { ...prev };

      switch (filterType) {
        case "search":
          delete newParams.q;
          break;
        case "category_id":
          delete newParams.category_id;
          break;
        case "status":
          delete newParams.status;
          break;
        case "priority":
          delete newParams.priority;
          break;
        case "tag_ids":
          delete newParams.tag_ids;
          delete newParams.tag_mode;
          break;
        case "date_range":
          delete newParams.due_date_from;
          delete newParams.due_date_to;
          break;
      }

      return { ...newParams, page: 1 };
    });
  }, []);

  // Calculate active filters
  const activeFilters = useMemo<ActiveFilters>(() => {
    const filters: ActiveFilters = {};

    if (searchParams.q) filters.search = searchParams.q;
    if (searchParams.category_id !== undefined) filters.category_id = searchParams.category_id as number | null;
    if (searchParams.status) filters.status = Array.isArray(searchParams.status) ? searchParams.status : [searchParams.status];
    if (searchParams.priority) filters.priority = Array.isArray(searchParams.priority) ? searchParams.priority : [searchParams.priority];
    if (searchParams.tag_ids?.length) filters.tag_ids = searchParams.tag_ids;
    if (searchParams.due_date_from || searchParams.due_date_to) {
      filters.date_range = {
        from: searchParams.due_date_from,
        to: searchParams.due_date_to,
      };
    }

    return filters;
  }, [searchParams.q, searchParams.category_id, searchParams.status, searchParams.priority, searchParams.tag_ids, searchParams.due_date_from, searchParams.due_date_to]);

  // Check if any filters are active
  const hasActiveFilters = useMemo(() => {
    return Object.keys(activeFilters).length > 0;
  }, [activeFilters]);

  return {
    searchParams,
    activeFilters,
    hasActiveFilters,
    updateSearchQuery,
    updateCategory,
    updateStatus,
    updatePriority,
    updateTags,
    updateDateRange,
    updateSort,
    updatePage,
    clearFilters,
    clearSingleFilter,
  };
}
