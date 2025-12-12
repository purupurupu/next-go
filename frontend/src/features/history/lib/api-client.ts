import { HttpClient } from "@/lib/api-client";
import { TodoHistory } from "../types/history";

export class HistoryApiClient extends HttpClient {
  async getHistories(todoId: number): Promise<TodoHistory[]> {
    return this.get<TodoHistory[]>(`/api/v1/todos/${todoId}/histories`);
  }
}

export const historyApiClient = new HistoryApiClient();
