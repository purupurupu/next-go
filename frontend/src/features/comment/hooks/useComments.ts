import { useState, useCallback, useEffect } from "react";
import { Comment, CreateCommentData, UpdateCommentData } from "../types/comment";
import { commentApiClient } from "../lib/api-client";
import { toast } from "sonner";

export function useComments(todoId: number | null) {
  const [comments, setComments] = useState<Comment[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // コメント一覧の取得
  const fetchComments = useCallback(async () => {
    if (!todoId) return;

    setIsLoading(true);
    setError(null);

    try {
      const data = await commentApiClient.getComments(todoId);
      setComments(data);
    } catch (err) {
      const message = err instanceof Error ? err.message : "コメントの取得に失敗しました";
      setError(message);
      toast.error(message);
    } finally {
      setIsLoading(false);
    }
  }, [todoId]);

  // コメントの作成
  const createComment = useCallback(async (data: CreateCommentData) => {
    if (!todoId) return;

    try {
      const newComment = await commentApiClient.createComment(todoId, data);
      setComments((prev) => [...prev, newComment]);
      toast.success("コメントを追加しました");
      return newComment;
    } catch (err) {
      const message = err instanceof Error ? err.message : "コメントの作成に失敗しました";
      toast.error(message);
      throw err;
    }
  }, [todoId]);

  // コメントの更新
  const updateComment = useCallback(async (commentId: number, data: UpdateCommentData) => {
    if (!todoId) return;

    try {
      const updatedComment = await commentApiClient.updateComment(todoId, commentId, data);
      setComments((prev) => prev.map((comment) =>
        comment.id === commentId ? updatedComment : comment,
      ));
      toast.success("コメントを更新しました");
      return updatedComment;
    } catch (err) {
      const message = err instanceof Error ? err.message : "コメントの更新に失敗しました";
      toast.error(message);
      throw err;
    }
  }, [todoId]);

  // コメントの削除
  const deleteComment = useCallback(async (commentId: number) => {
    if (!todoId) return;

    try {
      await commentApiClient.deleteComment(todoId, commentId);
      setComments((prev) => prev.filter((comment) => comment.id !== commentId));
      toast.success("コメントを削除しました");
    } catch (err) {
      const message = err instanceof Error ? err.message : "コメントの削除に失敗しました";
      toast.error(message);
      throw err;
    }
  }, [todoId]);

  // todoIdが変更されたらコメントを再取得
  useEffect(() => {
    if (todoId) {
      fetchComments();
    } else {
      setComments([]);
    }
  }, [todoId, fetchComments]);

  return {
    comments,
    isLoading,
    error,
    fetchComments,
    createComment,
    updateComment,
    deleteComment,
  };
}
