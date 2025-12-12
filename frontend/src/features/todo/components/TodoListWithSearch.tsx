"use client";

import { useState } from "react";
import { Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";

import { useSearchParams } from "../hooks/useSearchParams";
import { useTodoSearch } from "../hooks/useTodoSearch";
import { useTodoMutations } from "../hooks/useTodoMutations";
import { useCategories } from "@/features/category/hooks/useCategories";
import { useTags } from "@/features/tag/hooks/useTags";

import { TodoItem } from "./TodoItem";
import { TodoForm } from "./TodoForm";
import { SearchBar } from "./SearchBar";
import { AdvancedFilters } from "./AdvancedFilters";
import { FilterBadges } from "./FilterBadges";

import type { Todo, CreateTodoData, UpdateTodoData } from "../types/todo";

export function TodoListWithSearch() {
  const [isCreateFormOpen, setIsCreateFormOpen] = useState(false);
  const [editingTodo, setEditingTodo] = useState<Todo | null>(null);

  // Search params management
  const {
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
  } = useSearchParams();

  // Fetch todos with search
  const { todos, loading, error, searchResponse, refreshSearch } = useTodoSearch(searchParams);

  // Categories and Tags for filters
  const { categories } = useCategories();
  const { tags } = useTags();

  // Todo mutations
  const mutations = useTodoMutations({
    allTodos: todos,
    setAllTodos: () => {}, // We'll refresh search instead
    setError: () => {},
  });

  const handleCreateTodo = async (data: CreateTodoData, files?: File[]) => {
    await mutations.createTodo(data, files);
    await refreshSearch();
  };

  const handleUpdateTodo = async (data: UpdateTodoData, files?: File[]) => {
    if (!editingTodo) return;
    await mutations.updateTodo(editingTodo.id, data, files);
    setEditingTodo(null);
    await refreshSearch();
  };

  const handleDeleteTodo = async (id: number) => {
    await mutations.deleteTodo(id);
    await refreshSearch();
  };

  const handleToggleComplete = async (id: number) => {
    await mutations.toggleTodoComplete(id);
    await refreshSearch();
  };

  const handleDeleteFile = async (todoId: number, fileId: string | number) => {
    await mutations.deleteTodoFile(todoId, fileId);
    await refreshSearch();
  };

  if (loading && !todos.length) {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold">TODO</h1>
          <Button disabled>
            <Plus className="h-4 w-4 mr-2" />
            タスクを追加
          </Button>
        </div>
        <Skeleton className="h-10 w-full" />
        <div className="space-y-3">
          {[...Array(5)].map((_, i) => (
            <Skeleton key={i} className="h-16 w-full" />
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

      {/* Search and Filters */}
      <div className="space-y-4">
        <SearchBar
          value={searchParams.q || ""}
          onChange={updateSearchQuery}
          placeholder="タイトルや説明で検索..."
        />

        <AdvancedFilters
          searchParams={searchParams}
          categories={categories || []}
          tags={tags || []}
          onUpdateCategory={updateCategory}
          onUpdateStatus={updateStatus}
          onUpdatePriority={updatePriority}
          onUpdateTags={updateTags}
          onUpdateDateRange={updateDateRange}
          onUpdateSort={updateSort}
        />

        {hasActiveFilters && (
          <FilterBadges
            activeFilters={activeFilters}
            categories={categories || []}
            tags={tags || []}
            onRemoveFilter={clearSingleFilter}
            onClearAll={clearFilters}
          />
        )}
      </div>

      <Separator />

      {/* Results Summary */}
      {searchResponse && (
        <div className="flex items-center justify-between text-sm text-muted-foreground">
          <span>
            {searchResponse.meta.total}
            {" "}
            件の結果
            {searchResponse.meta.total_pages > 1 && ` (${searchResponse.meta.current_page}/${searchResponse.meta.total_pages} ページ)`}
          </span>
          {loading && <span>更新中...</span>}
        </div>
      )}

      {/* Todo List */}
      <div className="space-y-3">
        {todos.length === 0
          ? (
              <div className="text-center py-12 text-muted-foreground">
                {hasActiveFilters
                  ? (
                      <div className="space-y-4">
                        <p>条件に一致するタスクが見つかりませんでした</p>
                        {searchResponse?.suggestions && searchResponse.suggestions.length > 0 && (
                          <div className="space-y-2">
                            <p className="text-sm font-medium">提案:</p>
                            {searchResponse.suggestions.slice(0, 3).map((suggestion, index) => (
                              <p key={index} className="text-sm">{suggestion.message}</p>
                            ))}
                          </div>
                        )}
                        <Button variant="outline" size="sm" onClick={clearFilters}>
                          フィルターをクリア
                        </Button>
                      </div>
                    )
                  : (
                      <div className="space-y-2">
                        <p>まだタスクがありません</p>
                        <p className="text-sm">「タスクを追加」ボタンから新しいタスクを作成しましょう</p>
                      </div>
                    )}
              </div>
            )
          : (
              todos.map((todo) => (
                <TodoItem
                  key={todo.id}
                  todo={todo}
                  onToggleComplete={handleToggleComplete}
                  onEdit={setEditingTodo}
                  onDelete={handleDeleteTodo}
                />
              ))
            )}
      </div>

      {/* Pagination */}
      {searchResponse && searchResponse.meta.total_pages > 1 && (
        <div className="flex justify-center gap-2 pt-4">
          <Button
            variant="outline"
            size="sm"
            disabled={searchResponse.meta.current_page === 1}
            onClick={() => updatePage(searchResponse.meta.current_page - 1)}
          >
            前のページ
          </Button>
          <span className="flex items-center px-3 text-sm">
            {searchResponse.meta.current_page}
            {" "}
            /
            {searchResponse.meta.total_pages}
          </span>
          <Button
            variant="outline"
            size="sm"
            disabled={searchResponse.meta.current_page === searchResponse.meta.total_pages}
            onClick={() => updatePage(searchResponse.meta.current_page + 1)}
          >
            次のページ
          </Button>
        </div>
      )}

      {/* Forms */}
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
          onFileDelete={(fileId) => handleDeleteFile(editingTodo.id, fileId)}
        />
      )}
    </div>
  );
}
