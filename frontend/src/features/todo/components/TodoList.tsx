"use client";

import { useState } from "react";
import { Plus } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";

import { useTodos } from "@/features/todo/hooks/useTodos";
import { TodoItem } from "./TodoItem";
import { TodoForm } from "./TodoForm";
import { TodoFilters } from "./TodoFilters";

import type { Todo, CreateTodoData, UpdateTodoData } from "@/features/todo/types/todo";

export function TodoList() {
  const {
    todos,
    allTodos,
    loading,
    error,
    filter,
    createTodo,
    updateTodo,
    deleteTodo,
    toggleTodoComplete,
    setFilter,
    deleteTodoFile,
  } = useTodos();

  const [isCreateFormOpen, setIsCreateFormOpen] = useState(false);
  const [editingTodo, setEditingTodo] = useState<Todo | null>(null);

  // Calculate counts based on all todos (not filtered)
  const counts = {
    all: allTodos.length,
    active: allTodos.filter((todo) => !todo.completed).length,
    completed: allTodos.filter((todo) => todo.completed).length,
  };

  const handleCreateTodo = async (data: CreateTodoData, files?: File[]) => {
    await createTodo(data, files);
  };

  const handleUpdateTodo = async (data: UpdateTodoData, files?: File[]) => {
    if (!editingTodo) return;
    await updateTodo(editingTodo.id, data, files);
    setEditingTodo(null);
  };

  const handleEditTodo = (todo: Todo) => {
    setEditingTodo(todo);
  };

  const handleDeleteTodo = async (id: number) => {
    await deleteTodo(id);
  };

  if (loading) {
    return (
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold">TODO</h1>
          <Button disabled>
            <Plus className="h-4 w-4 mr-2" />
            タスクを追加
          </Button>
        </div>
        <div className="space-y-3">
          {[...Array(3)].map((_, i) => (
            <div
              key={i}
              className="h-16 bg-muted animate-pulse rounded-lg"
            />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">TODO</h1>
        <Button onClick={() => setIsCreateFormOpen(true)}>
          <Plus className="h-4 w-4 mr-2" />
          タスクを追加
        </Button>
      </div>

      {error && (
        <div className="p-4 text-sm text-destructive bg-destructive/10 border border-destructive/20 rounded-lg">
          {error}
        </div>
      )}

      <TodoFilters
        currentFilter={filter}
        onFilterChange={setFilter}
        counts={counts}
      />

      <Separator />

      <div className="space-y-3">
        {todos.length === 0
          ? (
              <div className="text-center py-12 text-muted-foreground">
                {filter === "all"
                  ? (
                      <div className="space-y-2">
                        <p>まだタスクがありません</p>
                        <p className="text-sm">「タスクを追加」ボタンから新しいタスクを作成しましょう</p>
                      </div>
                    )
                  : filter === "active"
                    ? (
                        <p>未完了のタスクはありません</p>
                      )
                    : (
                        <p>完了したタスクはありません</p>
                      )}
              </div>
            )
          : (
              todos.map((todo) => (
                <TodoItem
                  key={todo.id}
                  todo={todo}
                  onToggleComplete={toggleTodoComplete}
                  onEdit={handleEditTodo}
                  onDelete={handleDeleteTodo}
                />
              ))
            )}
      </div>

      {isCreateFormOpen && (
        <TodoForm
          mode="create"
          open={isCreateFormOpen}
          onOpenChange={setIsCreateFormOpen}
          onSubmit={handleCreateTodo}
        />
      )}

      {editingTodo && (
        <TodoForm
          mode="edit"
          todo={editingTodo}
          open={!!editingTodo}
          onOpenChange={(open) => !open && setEditingTodo(null)}
          onSubmit={handleUpdateTodo}
          onFileDelete={(fileId) => deleteTodoFile(editingTodo.id, fileId)}
        />
      )}
    </div>
  );
}
