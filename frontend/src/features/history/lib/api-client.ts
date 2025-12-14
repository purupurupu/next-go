import { ApiClient } from "@/lib/api-client";
import { TodoHistory } from "../types/history";

export class HistoryApiClient extends ApiClient {
  async getHistories(todoId: number): Promise<TodoHistory[]> {
    return this.get<TodoHistory[]>(`/todos/${todoId}/histories`);
  }
}

export const historyApiClient = new HistoryApiClient();
