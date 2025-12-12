"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import type { CreateCategoryData, UpdateCategoryData, Category } from "../types/category";

interface CategoryFormProps {
  category?: Category;
  onSubmit: (data: CreateCategoryData | UpdateCategoryData) => Promise<void>;
  onCancel: () => void;
}

const defaultColors = [
  "#EF4444", // red
  "#F59E0B", // amber
  "#10B981", // emerald
  "#3B82F6", // blue
  "#8B5CF6", // violet
  "#EC4899", // pink
  "#6B7280", // gray
  "#059669", // green
];

export function CategoryForm({ category, onSubmit, onCancel }: CategoryFormProps) {
  const [name, setName] = useState(category?.name || "");
  const [color, setColor] = useState(category?.color || "#3B82F6");
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;

    setIsSubmitting(true);
    try {
      await onSubmit({ name: name.trim(), color });
      if (!category) {
        setName("");
        setColor("#3B82F6");
      }
    } catch {
      // エラーは親コンポーネントで処理
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="category-name">カテゴリー名</Label>
        <Input
          id="category-name"
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="仕事、プライベートなど"
          required
          maxLength={50}
          disabled={isSubmitting}
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor="category-color">カラー</Label>
        <div className="flex items-center gap-2">
          <Input
            id="category-color"
            type="color"
            value={color}
            onChange={(e) => setColor(e.target.value)}
            className="h-10 w-20"
            disabled={isSubmitting}
          />
          <div className="flex gap-1">
            {defaultColors.map((defaultColor) => (
              <button
                key={defaultColor}
                type="button"
                onClick={() => setColor(defaultColor)}
                className="h-8 w-8 rounded border-2 border-gray-200 hover:border-gray-400"
                style={{ backgroundColor: defaultColor }}
                disabled={isSubmitting}
              />
            ))}
          </div>
        </div>
      </div>

      <div className="flex justify-end gap-2">
        <Button
          type="button"
          variant="outline"
          onClick={onCancel}
          disabled={isSubmitting}
        >
          キャンセル
        </Button>
        <Button type="submit" disabled={isSubmitting || !name.trim()}>
          {isSubmitting ? "保存中..." : category ? "更新" : "作成"}
        </Button>
      </div>
    </form>
  );
}
