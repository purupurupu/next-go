import { ApiClient } from "@/lib/api-client";
import { TodoHistory } from "../types/history";

interface HistoryListResponse {
  histories: TodoHistory[];
  meta: {
    total: number;
    current_page: number;
    total_pages: number;
    per_page: number;
  };
}

export class HistoryApiClient extends ApiClient {
  async getHistories(todoId: number): Promise<TodoHistory[]> {
    const response = await this.get<HistoryListResponse>(`/todos/${todoId}/histories`);
    return response.histories;
  }
}

export const historyApiClient = new HistoryApiClient();
