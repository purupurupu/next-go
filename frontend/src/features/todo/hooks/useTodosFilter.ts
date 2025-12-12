import { useMemo } from "react";
import type { Todo, TodoFilter } from "../types/todo";
import { TODO_FILTERS } from "@/lib/constants";

interface UseTodosFilterParams {
  allTodos: Todo[];
  filter: TodoFilter;
}

interface UseTodosFilterReturn {
  todos: Todo[];
  counts: {
    all: number;
    active: number;
    completed: number;
  };
}

/**
 * Hook for filtering and computing todo counts
 * Handles: filtering todos based on current filter, computing counts for all filter types
 */
export function useTodosFilter({ allTodos, filter }: UseTodosFilterParams): UseTodosFilterReturn {
  // Filter todos based on current filter
  const todos = useMemo(() => {
    return allTodos.filter((todo) => {
      switch (filter) {
        case TODO_FILTERS.ACTIVE:
          return !todo.completed;
        case TODO_FILTERS.COMPLETED:
          return todo.completed;
        default:
          return true;
      }
    });
  }, [allTodos, filter]);

  // Calculate counts for all filter types
  const counts = useMemo(() => ({
    all: allTodos.length,
    active: allTodos.filter((todo) => !todo.completed).length,
    completed: allTodos.filter((todo) => todo.completed).length,
  }), [allTodos]);

  return {
    todos,
    counts,
  };
}
