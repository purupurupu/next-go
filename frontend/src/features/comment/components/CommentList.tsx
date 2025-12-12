"use client";

// import { Comment } from "../types/comment";
import { CommentItem } from "./CommentItem";
import { CommentForm } from "./CommentForm";
import { useComments } from "../hooks/useComments";
import { Loader2 } from "lucide-react";

interface CommentListProps {
  todoId: number | null;
}

export function CommentList({ todoId }: CommentListProps) {
  const {
    comments,
    isLoading,
    createComment,
    updateComment,
    deleteComment,
  } = useComments(todoId);

  const handleCreateComment = async (content: string) => {
    await createComment({ content });
  };

  const handleUpdateComment = async (commentId: number, content: string) => {
    await updateComment(commentId, { content });
  };

  if (!todoId) {
    return null;
  }

  return (
    <div className="space-y-4">
      <div className="border-t pt-4">
        <h3 className="text-sm font-medium mb-3">コメント</h3>

        {isLoading
          ? (
              <div className="flex items-center justify-center py-8">
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
              </div>
            )
          : (
              <>
                {!comments || comments.length === 0
                  ? (
                      <p className="text-sm text-muted-foreground mb-4">
                        まだコメントはありません
                      </p>
                    )
                  : (
                      <div className="space-y-4 mb-4">
                        {Array.isArray(comments) && comments.map((comment) => (
                          <CommentItem
                            key={comment.id}
                            comment={comment}
                            onUpdate={handleUpdateComment}
                            onDelete={deleteComment}
                          />
                        ))}
                      </div>
                    )}

                <CommentForm onSubmit={handleCreateComment} />
              </>
            )}
      </div>
    </div>
  );
}
