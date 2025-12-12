import { useState, useCallback, useEffect } from "react";
import { TodoHistory } from "../types/history";
import { historyApiClient } from "../lib/api-client";
import { toast } from "sonner";

export function useHistory(todoId: number | null) {
  const [histories, setHistories] = useState<TodoHistory[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchHistories = useCallback(async () => {
    if (!todoId) return;

    setIsLoading(true);
    setError(null);

    try {
      const data = await historyApiClient.getHistories(todoId);
      setHistories(data);
    } catch (err) {
      const message = err instanceof Error ? err.message : "履歴の取得に失敗しました";
      setError(message);
      toast.error(message);
    } finally {
      setIsLoading(false);
    }
  }, [todoId]);

  // todoIdが変更されたら履歴を再取得
  useEffect(() => {
    if (todoId) {
      fetchHistories();
    } else {
      setHistories([]);
    }
  }, [todoId, fetchHistories]);

  return {
    histories,
    isLoading,
    error,
    fetchHistories,
  };
}
