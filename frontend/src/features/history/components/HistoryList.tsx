"use client";

import { useHistory } from "../hooks/useHistory";
import { HistoryItem } from "./HistoryItem";
import { Loader2, History } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useState } from "react";

interface HistoryListProps {
  todoId: number | null;
}

export function HistoryList({ todoId }: HistoryListProps) {
  const { histories, isLoading, error } = useHistory(todoId);
  const [isExpanded, setIsExpanded] = useState(false);

  if (!todoId) {
    return null;
  }

  const displayedHistories = isExpanded ? histories : histories.slice(0, 3);
  const hasMore = histories.length > 3;

  return (
    <div className="space-y-3">
      <div className="flex items-center gap-2">
        <History className="h-4 w-4 text-muted-foreground" />
        <h3 className="text-sm font-medium">変更履歴</h3>
      </div>

      {isLoading
        ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            </div>
          )
        : error
          ? (
              <p className="text-sm text-red-600">{error}</p>
            )
          : histories.length === 0
            ? (
                <p className="text-sm text-muted-foreground">
                  まだ履歴はありません
                </p>
              )
            : (
                <>
                  <div className="space-y-3">
                    {displayedHistories.map((history) => (
                      <HistoryItem key={history.id} history={history} />
                    ))}
                  </div>

                  {hasMore && (
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => setIsExpanded(!isExpanded)}
                      className="w-full"
                    >
                      {isExpanded ? "履歴を折りたたむ" : `他${histories.length - 3}件の履歴を表示`}
                    </Button>
                  )}
                </>
              )}
    </div>
  );
}
