import { generateOptimisticId } from "@/lib/utils";
import type { Todo, CreateTodoData, UpdateTodoData, UpdateOrderData } from "../types/todo";

/**
 * Creates an optimistic todo for immediate UI updates
 */
export function createOptimisticTodo(
  data: CreateTodoData,
  currentTodosLength: number,
): Todo {
  return {
    id: generateOptimisticId(),
    title: data.title,
    completed: false,
    position: currentTodosLength + 1,
    due_date: data.due_date || null,
    priority: data.priority || "medium",
    status: data.status || "pending",
    description: data.description || null,
    category: null, // Will be updated when real todo is returned from server
    tags: [], // Will be updated when real todo is returned from server
    files: [], // Will be updated when real todo is returned from server
    comments_count: 0,
    latest_comments: [],
    history_count: 0,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  };
}

/**
 * Applies optimistic update to todo list
 */
export function applyOptimisticUpdate(
  todos: Todo[],
  id: number,
  updates: Partial<UpdateTodoData>,
): Todo[] {
  return todos.map((todo) =>
    todo.id === id ? { ...todo, ...updates } : todo,
  );
}

/**
 * Applies optimistic order update to todo list
 */
export function applyOptimisticOrder(
  todos: Todo[],
  reorderedTodos: UpdateOrderData[],
): Todo[] {
  return [...todos].sort((a, b) => {
    const aData = reorderedTodos.find((item) => item.id === a.id);
    const bData = reorderedTodos.find((item) => item.id === b.id);
    return (aData?.position || 0) - (bData?.position || 0);
  });
}

/**
 * Adds optimistic todo to the end of the list
 */
export function addOptimisticTodo(todos: Todo[], newTodo: Todo): Todo[] {
  return [...todos, newTodo];
}

/**
 * Removes todo from list (for optimistic delete)
 */
export function removeOptimisticTodo(todos: Todo[], id: number): Todo[] {
  return todos.filter((todo) => todo.id !== id);
}

/**
 * Updates optimistic todo with real todo data from server
 */
export function updateOptimisticTodo(
  todos: Todo[],
  optimisticId: number,
  realTodo: Todo,
): Todo[] {
  return todos.map((todo) =>
    todo.id === optimisticId ? realTodo : todo,
  );
}
