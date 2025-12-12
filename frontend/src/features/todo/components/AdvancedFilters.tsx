"use client";

import { useState } from "react";
import { ChevronDown, ChevronUp, Filter } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Checkbox } from "@/components/ui/checkbox";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import { TagSelector } from "@/features/tag/components/TagSelector";
import { DatePicker } from "@/components/ui/date-picker";
import type { TodoSearchParams, TodoStatus, TodoPriority } from "../types/todo";
import type { Category } from "@/features/category/types/category";
import type { Tag } from "@/features/tag/types/tag";

interface AdvancedFiltersProps {
  searchParams: TodoSearchParams;
  categories: Category[];
  tags: Tag[];
  onUpdateCategory: (categoryId: number | null | undefined) => void;
  onUpdateStatus: (status: TodoStatus[]) => void;
  onUpdatePriority: (priority: TodoPriority[]) => void;
  onUpdateTags: (tagIds: number[], tagMode?: "any" | "all") => void;
  onUpdateDateRange: (from?: string, to?: string) => void;
  onUpdateSort: (sortBy: TodoSearchParams["sort_by"], sortOrder?: TodoSearchParams["sort_order"]) => void;
}

const statusOptions: Array<{ value: TodoStatus; label: string }> = [
  { value: "pending", label: "未着手" },
  { value: "in_progress", label: "進行中" },
  { value: "completed", label: "完了" },
];

const priorityOptions: Array<{ value: TodoPriority; label: string }> = [
  { value: "high", label: "高" },
  { value: "medium", label: "中" },
  { value: "low", label: "低" },
];

const sortOptions: Array<{ value: TodoSearchParams["sort_by"]; label: string }> = [
  { value: "position", label: "並び順" },
  { value: "created_at", label: "作成日" },
  { value: "updated_at", label: "更新日" },
  { value: "due_date", label: "期限" },
  { value: "title", label: "タイトル" },
  { value: "priority", label: "優先度" },
  { value: "status", label: "ステータス" },
];

export function AdvancedFilters({
  searchParams,
  categories,
  tags,
  onUpdateCategory,
  onUpdateStatus,
  onUpdatePriority,
  onUpdateTags,
  onUpdateDateRange,
  onUpdateSort,
}: AdvancedFiltersProps) {
  const [isOpen, setIsOpen] = useState(false);

  // Extract current values
  const currentStatus = Array.isArray(searchParams.status) ? searchParams.status : searchParams.status ? [searchParams.status] : [];
  const currentPriority = Array.isArray(searchParams.priority) ? searchParams.priority : searchParams.priority ? [searchParams.priority] : [];
  const currentTagIds = searchParams.tag_ids || [];
  const currentTagMode = searchParams.tag_mode || "any";

  const handleStatusChange = (status: TodoStatus, checked: boolean) => {
    const newStatus = checked
      ? [...currentStatus, status]
      : currentStatus.filter((s) => s !== status);
    onUpdateStatus(newStatus);
  };

  const handlePriorityChange = (priority: TodoPriority, checked: boolean) => {
    const newPriority = checked
      ? [...currentPriority, priority]
      : currentPriority.filter((p) => p !== priority);
    onUpdatePriority(newPriority);
  };

  const handleCategoryChange = (value: string) => {
    if (value === "all") {
      onUpdateCategory(undefined); // Clear filter
    } else if (value === "none") {
      onUpdateCategory(-1); // Backend expects -1 for uncategorized
    } else {
      onUpdateCategory(parseInt(value));
    }
  };

  const handleDateFromChange = (date: Date | undefined) => {
    onUpdateDateRange(date?.toISOString().split("T")[0], searchParams.due_date_to);
  };

  const handleDateToChange = (date: Date | undefined) => {
    onUpdateDateRange(searchParams.due_date_from, date?.toISOString().split("T")[0]);
  };

  const handleSortChange = (value: string) => {
    const [sortBy, sortOrder] = value.split("-") as [TodoSearchParams["sort_by"], TodoSearchParams["sort_order"]];
    onUpdateSort(sortBy, sortOrder);
  };

  return (
    <Collapsible open={isOpen} onOpenChange={setIsOpen}>
      <CollapsibleTrigger asChild>
        <Button variant="outline" size="sm" className="gap-2">
          <Filter className="h-4 w-4" />
          詳細フィルター
          {isOpen ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
        </Button>
      </CollapsibleTrigger>
      <CollapsibleContent className="space-y-4 pt-4">
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {/* Category Filter */}
          <div className="space-y-2">
            <Label>カテゴリー</Label>
            <Select
              value={searchParams.category_id === -1 || searchParams.category_id === null ? "none" : searchParams.category_id?.toString() || "all"}
              onValueChange={handleCategoryChange}
            >
              <SelectTrigger>
                <SelectValue placeholder="カテゴリーを選択" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">すべて</SelectItem>
                <SelectItem value="none">カテゴリーなし</SelectItem>
                {categories?.map((category) => (
                  <SelectItem key={category.id} value={category.id.toString()}>
                    <div className="flex items-center gap-2">
                      <div
                        className="h-3 w-3 rounded-full"
                        style={{ backgroundColor: category.color }}
                      />
                      {category.name}
                    </div>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Status Filter */}
          <div className="space-y-2">
            <Label>ステータス</Label>
            <div className="space-y-2 rounded-md border p-3">
              {statusOptions.map((option) => (
                <div key={option.value} className="flex items-center space-x-2">
                  <Checkbox
                    id={`status-${option.value}`}
                    checked={currentStatus.includes(option.value)}
                    onCheckedChange={(checked) => handleStatusChange(option.value, checked as boolean)}
                  />
                  <Label
                    htmlFor={`status-${option.value}`}
                    className="text-sm font-normal cursor-pointer"
                  >
                    {option.label}
                  </Label>
                </div>
              ))}
            </div>
          </div>

          {/* Priority Filter */}
          <div className="space-y-2">
            <Label>優先度</Label>
            <div className="space-y-2 rounded-md border p-3">
              {priorityOptions.map((option) => (
                <div key={option.value} className="flex items-center space-x-2">
                  <Checkbox
                    id={`priority-${option.value}`}
                    checked={currentPriority.includes(option.value)}
                    onCheckedChange={(checked) => handlePriorityChange(option.value, checked as boolean)}
                  />
                  <Label
                    htmlFor={`priority-${option.value}`}
                    className="text-sm font-normal cursor-pointer"
                  >
                    {option.label}
                  </Label>
                </div>
              ))}
            </div>
          </div>

          {/* Tag Filter */}
          <div className="space-y-2">
            <Label>タグ</Label>
            <TagSelector
              tags={tags}
              selectedTagIds={currentTagIds}
              onSelectionChange={(tagIds) => onUpdateTags(tagIds, currentTagMode)}
            />
            {currentTagIds.length > 0 && (
              <RadioGroup
                value={currentTagMode}
                onValueChange={(value: "any" | "all") => onUpdateTags(currentTagIds, value)}
                className="flex gap-4 pt-2"
              >
                <div className="flex items-center space-x-2">
                  <RadioGroupItem value="any" id="tag-any" />
                  <Label htmlFor="tag-any" className="text-sm font-normal cursor-pointer">
                    いずれか
                  </Label>
                </div>
                <div className="flex items-center space-x-2">
                  <RadioGroupItem value="all" id="tag-all" />
                  <Label htmlFor="tag-all" className="text-sm font-normal cursor-pointer">
                    すべて
                  </Label>
                </div>
              </RadioGroup>
            )}
          </div>

          {/* Date Range Filter */}
          <div className="space-y-2">
            <Label>期限（開始日）</Label>
            <DatePicker
              date={searchParams.due_date_from ? new Date(searchParams.due_date_from) : undefined}
              onSelect={handleDateFromChange}
              placeholder="開始日を選択"
            />
          </div>

          <div className="space-y-2">
            <Label>期限（終了日）</Label>
            <DatePicker
              date={searchParams.due_date_to ? new Date(searchParams.due_date_to) : undefined}
              onSelect={handleDateToChange}
              placeholder="終了日を選択"
            />
          </div>

          {/* Sort Options */}
          <div className="space-y-2">
            <Label>並び替え</Label>
            <Select
              value={`${searchParams.sort_by}-${searchParams.sort_order}`}
              onValueChange={handleSortChange}
            >
              <SelectTrigger>
                <SelectValue placeholder="並び順を選択" />
              </SelectTrigger>
              <SelectContent>
                {sortOptions.map((option) => (
                  <SelectItem key={`${option.value}-asc`} value={`${option.value}-asc`}>
                    {option.label}
                    {" "}
                    (昇順)
                  </SelectItem>
                ))}
                {sortOptions.map((option) => (
                  <SelectItem key={`${option.value}-desc`} value={`${option.value}-desc`}>
                    {option.label}
                    {" "}
                    (降順)
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>
      </CollapsibleContent>
    </Collapsible>
  );
}
