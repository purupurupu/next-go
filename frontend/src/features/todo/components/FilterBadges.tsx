import { X } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import type { ActiveFilters } from "../types/todo";
import type { Category } from "@/features/category/types/category";
import type { Tag } from "@/features/tag/types/tag";

interface FilterBadgesProps {
  activeFilters: ActiveFilters;
  categories?: Category[];
  tags?: Tag[];
  onRemoveFilter: (filterType: keyof ActiveFilters) => void;
  onClearAll: () => void;
}

export function FilterBadges({
  activeFilters,
  categories = [],
  tags = [],
  onRemoveFilter,
  onClearAll,
}: FilterBadgesProps) {
  const filterBadges: Array<{
    key: keyof ActiveFilters;
    label: string;
    value: string;
  }> = [];

  // Search filter
  if (activeFilters.search) {
    filterBadges.push({
      key: "search",
      label: "検索",
      value: activeFilters.search,
    });
  }

  // Category filter
  if (activeFilters.category_id !== undefined) {
    const category = categories.find((c) => c.id === activeFilters.category_id);
    filterBadges.push({
      key: "category_id",
      label: "カテゴリー",
      value: activeFilters.category_id === null ? "カテゴリーなし" : category?.name || "Unknown",
    });
  }

  // Status filter
  if (activeFilters.status?.length) {
    const statusLabels = {
      pending: "未着手",
      in_progress: "進行中",
      completed: "完了",
    };
    filterBadges.push({
      key: "status",
      label: "ステータス",
      value: activeFilters.status.map((s) => statusLabels[s]).join(", "),
    });
  }

  // Priority filter
  if (activeFilters.priority?.length) {
    const priorityLabels = {
      low: "低",
      medium: "中",
      high: "高",
    };
    filterBadges.push({
      key: "priority",
      label: "優先度",
      value: activeFilters.priority.map((p) => priorityLabels[p]).join(", "),
    });
  }

  // Tag filter
  if (activeFilters.tag_ids?.length) {
    const selectedTags = tags.filter((t) => activeFilters.tag_ids?.includes(t.id));
    filterBadges.push({
      key: "tag_ids",
      label: "タグ",
      value: selectedTags.map((t) => t.name).join(", "),
    });
  }

  // Date range filter
  if (activeFilters.date_range?.from || activeFilters.date_range?.to) {
    const from = activeFilters.date_range.from ? new Date(activeFilters.date_range.from).toLocaleDateString("ja-JP") : "";
    const to = activeFilters.date_range.to ? new Date(activeFilters.date_range.to).toLocaleDateString("ja-JP") : "";

    let value = "";
    if (from && to) {
      value = `${from} - ${to}`;
    } else if (from) {
      value = `${from} 以降`;
    } else if (to) {
      value = `${to} まで`;
    }

    filterBadges.push({
      key: "date_range",
      label: "期限",
      value,
    });
  }

  if (filterBadges.length === 0) {
    return null;
  }

  return (
    <div className="flex flex-wrap items-center gap-2">
      <span className="text-sm text-muted-foreground">フィルター:</span>
      {filterBadges.map((badge) => (
        <Badge
          key={badge.key}
          variant="secondary"
          className="gap-1 pr-1"
        >
          <span className="text-xs">
            {badge.label}
            :
            {badge.value}
          </span>
          <Button
            variant="ghost"
            size="sm"
            className="h-4 w-4 p-0 hover:bg-transparent"
            onClick={() => onRemoveFilter(badge.key)}
          >
            <X className="h-3 w-3" />
          </Button>
        </Badge>
      ))}
      <Button
        variant="ghost"
        size="sm"
        onClick={onClearAll}
        className="h-7 px-2 text-xs"
      >
        すべてクリア
      </Button>
    </div>
  );
}
