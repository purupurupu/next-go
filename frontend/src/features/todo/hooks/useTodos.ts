import { useEffect } from "react";
import { useTodosState } from "./useTodosState";
import { useTodoMutations } from "./useTodoMutations";
import { useTodosFilter } from "./useTodosFilter";
import type { Todo, CreateTodoData, UpdateTodoData, UpdateOrderData, TodoFilter } from "../types/todo";

interface UseTodosReturn {
  // State
  todos: Todo[];
  allTodos: Todo[];
  loading: boolean;
  error: string | null;
  filter: TodoFilter;
  counts: {
    all: number;
    active: number;
    completed: number;
  };
  // Actions
  createTodo: (data: CreateTodoData, files?: File[]) => Promise<void>;
  updateTodo: (id: number, data: UpdateTodoData, files?: File[]) => Promise<void>;
  deleteTodo: (id: number) => Promise<void>;
  updateTodoOrder: (todos: UpdateOrderData[]) => Promise<void>;
  toggleTodoComplete: (id: number) => Promise<void>;
  setFilter: (filter: TodoFilter) => void;
  refreshTodos: () => Promise<void>;
  deleteTodoFile: (todoId: number, fileId: string | number) => Promise<void>;
}

/**
 * Main todos hook that combines all todo-related functionality
 * Integrates: state management, CRUD operations, filtering, and initial loading
 */
export function useTodos(): UseTodosReturn {
  // State management
  const {
    allTodos,
    loading,
    error,
    filter,
    setAllTodos,
    setLoading,
    setError,
    setFilter,
  } = useTodosState();

  // CRUD operations with optimistic updates
  const mutations = useTodoMutations({
    allTodos,
    setAllTodos,
    setError,
  });

  // Filtering and counts
  const { todos, counts } = useTodosFilter({ allTodos, filter });

  // Enhanced refresh function with loading state
  const refreshTodos = async () => {
    setLoading(true);
    try {
      await mutations.refreshTodos();
    } finally {
      setLoading(false);
    }
  };

  // Initial load
  useEffect(() => {
    refreshTodos();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []); // We want this to run only once on mount

  return {
    // State
    todos,
    allTodos,
    loading,
    error,
    filter,
    counts,
    // Actions
    createTodo: mutations.createTodo,
    updateTodo: mutations.updateTodo,
    deleteTodo: mutations.deleteTodo,
    updateTodoOrder: mutations.updateTodoOrder,
    toggleTodoComplete: mutations.toggleTodoComplete,
    setFilter,
    refreshTodos,
    deleteTodoFile: mutations.deleteTodoFile,
  };
}
