import { HttpClient } from "@/lib/api-client";
import { Comment, CreateCommentData, UpdateCommentData } from "../types/comment";

export class CommentApiClient extends HttpClient {
  async getComments(todoId: number): Promise<Comment[]> {
    const response = await this.get<Comment[]>(`/api/v1/todos/${todoId}/comments`);
    // 配列であることを保証
    return Array.isArray(response) ? response : [];
  }

  async createComment(todoId: number, data: CreateCommentData): Promise<Comment> {
    return this.post<Comment>(`/api/v1/todos/${todoId}/comments`, { comment: data });
  }

  async updateComment(todoId: number, commentId: number, data: UpdateCommentData): Promise<Comment> {
    return this.patch<Comment>(`/api/v1/todos/${todoId}/comments/${commentId}`, { comment: data });
  }

  async deleteComment(todoId: number, commentId: number): Promise<void> {
    await this.delete(`/api/v1/todos/${todoId}/comments/${commentId}`);
  }
}

export const commentApiClient = new CommentApiClient();
