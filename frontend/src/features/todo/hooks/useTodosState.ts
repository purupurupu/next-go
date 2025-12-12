import { useState, useCallback } from "react";
import type { Todo, TodoFilter } from "../types/todo";
import { TODO_FILTERS } from "@/lib/constants";

interface UseTodosState {
  allTodos: Todo[];
  loading: boolean;
  error: string | null;
  filter: TodoFilter;
}

interface UseTodosStateActions {
  setAllTodos: (todos: Todo[] | ((prev: Todo[]) => Todo[])) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  setFilter: (filter: TodoFilter) => void;
}

interface UseTodosStateReturn extends UseTodosState, UseTodosStateActions {}

/**
 * Hook for managing todos state
 * Handles: state management, setters, computed values
 */
export function useTodosState(): UseTodosStateReturn {
  const [state, setState] = useState<UseTodosState>({
    allTodos: [],
    loading: false,
    error: null,
    filter: TODO_FILTERS.ALL,
  });

  const setLoading = useCallback((loading: boolean) => {
    setState((prev) => ({ ...prev, loading }));
  }, []);

  const setError = useCallback((error: string | null) => {
    setState((prev) => ({ ...prev, error }));
  }, []);

  const setAllTodos = useCallback((todos: Todo[] | ((prev: Todo[]) => Todo[])) => {
    setState((prev) => ({
      ...prev,
      allTodos: typeof todos === "function" ? todos(prev.allTodos) : todos,
    }));
  }, []);

  const setFilter = useCallback((filter: TodoFilter) => {
    setState((prev) => ({ ...prev, filter }));
  }, []);

  return {
    ...state,
    setAllTodos,
    setLoading,
    setError,
    setFilter,
  };
}
