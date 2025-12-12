"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Pencil, Trash2 } from "lucide-react";
import type { Category } from "../types/category";
import { CategoryForm } from "./CategoryForm";

interface CategoryListProps {
  categories: Category[];
  onUpdate: (id: number, data: { name: string; color: string }) => Promise<void>;
  onDelete: (id: number) => Promise<void>;
}

export function CategoryList({ categories, onUpdate, onDelete }: CategoryListProps) {
  const [editingCategory, setEditingCategory] = useState<Category | null>(null);

  const handleDelete = async (id: number) => {
    if (confirm("このカテゴリーを削除しますか？関連するTodoからカテゴリーが外されます。")) {
      await onDelete(id);
    }
  };

  if (editingCategory) {
    return (
      <Card className="p-4">
        <h3 className="mb-4 text-lg font-semibold">カテゴリーを編集</h3>
        <CategoryForm
          category={editingCategory}
          onSubmit={async (data) => {
            await onUpdate(editingCategory.id, data as { name: string; color: string });
            setEditingCategory(null);
          }}
          onCancel={() => setEditingCategory(null)}
        />
      </Card>
    );
  }

  if (categories.length === 0) {
    return (
      <div className="text-center text-gray-500 py-8">
        カテゴリーがまだ作成されていません
      </div>
    );
  }

  return (
    <div className="space-y-2">
      {categories.map((category) => (
        <Card
          key={category.id}
          className="flex items-center justify-between p-4 hover:bg-gray-50"
        >
          <div className="flex items-center gap-3">
            <div
              className="h-6 w-6 rounded"
              style={{ backgroundColor: category.color }}
            />
            <div>
              <p className="font-medium">{category.name}</p>
              <p className="text-sm text-gray-500">
                {category.todo_count}
                {" "}
                個のタスク
              </p>
            </div>
          </div>
          <div className="flex gap-2">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setEditingCategory(category)}
            >
              <Pencil className="h-4 w-4" />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => handleDelete(category.id)}
              className="text-red-600 hover:text-red-700 hover:bg-red-50"
            >
              <Trash2 className="h-4 w-4" />
            </Button>
          </div>
        </Card>
      ))}
    </div>
  );
}
