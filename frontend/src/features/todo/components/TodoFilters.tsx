import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

import { TODO_FILTERS } from "@/lib/constants";
import type { TodoFilter } from "@/features/todo/types/todo";

interface TodoFiltersProps {
  currentFilter: TodoFilter;
  onFilterChange: (filter: TodoFilter) => void;
  counts: {
    all: number;
    active: number;
    completed: number;
  };
}

export function TodoFilters({ currentFilter, onFilterChange, counts }: TodoFiltersProps) {
  const filters = [
    {
      key: TODO_FILTERS.ALL,
      label: "すべて",
      count: counts.all,
    },
    {
      key: TODO_FILTERS.ACTIVE,
      label: "未完了",
      count: counts.active,
    },
    {
      key: TODO_FILTERS.COMPLETED,
      label: "完了済み",
      count: counts.completed,
    },
  ] as const;

  return (
    <div className="flex flex-wrap gap-2">
      {filters.map((filter) => (
        <Button
          key={filter.key}
          variant={currentFilter === filter.key ? "default" : "outline"}
          size="sm"
          onClick={() => onFilterChange(filter.key)}
          className={cn(
            "flex items-center gap-2",
            currentFilter === filter.key && "pointer-events-none",
          )}
        >
          {filter.label}
          <Badge
            variant={currentFilter === filter.key ? "secondary" : "outline"}
            className="text-xs"
          >
            {filter.count}
          </Badge>
        </Button>
      ))}
    </div>
  );
}
