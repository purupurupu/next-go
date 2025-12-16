"use client";

import { useState } from "react";
import { format } from "date-fns";
import { ja } from "date-fns/locale";
import { Calendar, Clock, Edit, Trash2, ChevronDown, ChevronUp } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { isOverdue, isDueToday, isDueSoon } from "@/lib/utils";

import type { Todo } from "@/features/todo/types/todo";
import { TagBadge } from "@/features/tag/components/TagBadge";
import { HighlightedText } from "@/features/todo/components/HighlightedText";

interface TodoItemProps {
  todo: Todo;
  onToggleComplete: (id: number) => void;
  onEdit: (todo: Todo) => void;
  onDelete: (id: number) => void;
}

export function TodoItem({ todo, onToggleComplete, onEdit, onDelete }: TodoItemProps) {
  const [isDeleting, setIsDeleting] = useState(false);
  const [showDescription, setShowDescription] = useState(false);

  const handleDelete = () => {
    setIsDeleting(true);
    try {
      onDelete(todo.id);
    } finally {
      setIsDeleting(false);
    }
  };

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case "high":
        return "destructive";
      case "medium":
        return "default";
      case "low":
        return "secondary";
      default:
        return "outline";
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case "completed":
        return "default";
      case "in_progress":
        return "secondary";
      case "pending":
        return "outline";
      default:
        return "outline";
    }
  };

  const getPriorityLabel = (priority: string) => {
    switch (priority) {
      case "high":
        return "高";
      case "medium":
        return "中";
      case "low":
        return "低";
      default:
        return priority;
    }
  };

  const getStatusLabel = (status: string) => {
    switch (status) {
      case "completed":
        return "完了";
      case "in_progress":
        return "進行中";
      case "pending":
        return "未着手";
      default:
        return status;
    }
  };

  const getDueDateStatus = () => {
    if (!todo.due_date) return null;

    if (isOverdue(todo.due_date)) {
      return { variant: "destructive" as const, label: "期限切れ", icon: Clock };
    }

    if (isDueToday(todo.due_date)) {
      return { variant: "default" as const, label: "今日まで", icon: Calendar };
    }

    if (isDueSoon(todo.due_date)) {
      return { variant: "secondary" as const, label: "期限間近", icon: Calendar };
    }

    return { variant: "outline" as const, label: format(new Date(todo.due_date), "M/d", { locale: ja }), icon: Calendar };
  };

  const dueDateStatus = getDueDateStatus();

  return (
    <Card className={cn(
      "transition-all duration-200 hover:shadow-md",
      todo.completed && "opacity-60",
      isDeleting && "opacity-50 pointer-events-none",
    )}
    >
      <CardContent className="p-4">
        <div className="flex items-start gap-3 w-full">
          <Checkbox
            checked={todo.completed}
            onCheckedChange={() => onToggleComplete(todo.id)}
            className="mt-0.5 flex-shrink-0"
          />

          <div className="flex-1 min-w-0 w-full">
            <div className="flex items-start justify-between gap-2">
              <h3 className={cn(
                "text-sm font-medium break-words flex-1 min-w-0",
                todo.completed && "line-through text-muted-foreground",
              )}
              >
                <HighlightedText
                  text={todo.title}
                  highlights={todo.highlights?.title}
                />
              </h3>

              <div className="flex items-center gap-1 flex-shrink-0">
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={() => onEdit(todo)}
                  className="h-8 w-8"
                >
                  <Edit className="h-3 w-3" />
                  <span className="sr-only">編集</span>
                </Button>

                <Button
                  variant="ghost"
                  size="icon"
                  onClick={handleDelete}
                  disabled={isDeleting}
                  className="h-8 w-8 text-destructive hover:text-destructive"
                >
                  <Trash2 className="h-3 w-3" />
                  <span className="sr-only">削除</span>
                </Button>
              </div>
            </div>

            <div className="flex items-center gap-2 mt-2 flex-wrap">
              {todo.category && (
                <Badge
                  variant="outline"
                  className="text-xs"
                  style={{
                    borderColor: todo.category.color,
                    backgroundColor: `${todo.category.color}20`,
                  }}
                >
                  <div
                    className="h-2 w-2 rounded-full mr-1"
                    style={{ backgroundColor: todo.category.color }}
                  />
                  {todo.category.name}
                </Badge>
              )}
              <Badge variant={getPriorityColor(todo.priority) as "destructive" | "default" | "secondary" | "outline"} className="text-xs">
                優先度:
                {" "}
                {getPriorityLabel(todo.priority)}
              </Badge>
              <Badge variant={getStatusColor(todo.status) as "destructive" | "default" | "secondary" | "outline"} className="text-xs">
                {getStatusLabel(todo.status)}
              </Badge>
              {dueDateStatus && (
                <Badge variant={dueDateStatus.variant} className="text-xs">
                  <dueDateStatus.icon className="h-3 w-3 mr-1" />
                  {dueDateStatus.label}
                </Badge>
              )}
            </div>

            {todo.tags && todo.tags.length > 0 && (
              <div className="flex items-center gap-1 mt-2 flex-wrap">
                {todo.tags.map((tag) => (
                  <TagBadge
                    key={tag.id}
                    name={tag.name}
                    color={tag.color}
                    className="text-xs"
                  />
                ))}
              </div>
            )}

            {todo.description && (
              <div className="mt-2">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setShowDescription(!showDescription)}
                  className="h-auto p-0 font-normal text-xs text-muted-foreground hover:text-foreground"
                >
                  {showDescription
                    ? (
                        <>
                          <ChevronUp className="h-3 w-3 mr-1" />
                          説明を隠す
                        </>
                      )
                    : (
                        <>
                          <ChevronDown className="h-3 w-3 mr-1" />
                          説明を表示
                        </>
                      )}
                </Button>
                {showDescription && (
                  <div className="mt-1 p-2 bg-muted/50 rounded text-xs text-muted-foreground whitespace-pre-wrap">
                    <HighlightedText
                      text={todo.description}
                      highlights={todo.highlights?.description}
                    />
                  </div>
                )}
              </div>
            )}

          </div>
        </div>
      </CardContent>
    </Card>
  );
}
