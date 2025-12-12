import { HttpClient, ApiError } from "@/lib/api-client";
import { API_ENDPOINTS } from "@/lib/constants";
import type {
  Todo,
  CreateTodoData,
  UpdateTodoData,
  UpdateOrderData,
  TodoSearchParams,
  TodoSearchResponse,
} from "@/features/todo/types/todo";

class TodoApiClient extends HttpClient {
  async getTodos(): Promise<Todo[]> {
    const response = await this.get<Todo[]>(API_ENDPOINTS.TODOS);
    // 配列であることを保証
    return Array.isArray(response) ? response : [];
  }

  async getTodoById(id: number): Promise<Todo> {
    return this.get<Todo>(API_ENDPOINTS.TODO_BY_ID(id));
  }

  async searchTodos(params: TodoSearchParams): Promise<TodoSearchResponse> {
    // Build query string from params
    const queryParams = new URLSearchParams();

    // Add search query
    if (params.q) queryParams.append("q", params.q);

    // Add category filter
    if (params.category_id !== undefined) {
      if (Array.isArray(params.category_id)) {
        params.category_id.forEach((id) => queryParams.append("category_id[]", String(id)));
      } else if (params.category_id === null) {
        queryParams.append("category_id", "-1"); // Backend expects -1 for uncategorized
      } else {
        queryParams.append("category_id", String(params.category_id));
      }
    }

    // Add status filter
    if (params.status) {
      if (Array.isArray(params.status)) {
        params.status.forEach((status) => queryParams.append("status[]", status));
      } else {
        queryParams.append("status", params.status);
      }
    }

    // Add priority filter
    if (params.priority) {
      if (Array.isArray(params.priority)) {
        params.priority.forEach((priority) => queryParams.append("priority[]", priority));
      } else {
        queryParams.append("priority", params.priority);
      }
    }

    // Add tag filters
    if (params.tag_ids?.length) {
      params.tag_ids.forEach((id) => queryParams.append("tag_ids[]", String(id)));
    }
    if (params.tag_mode) queryParams.append("tag_mode", params.tag_mode);

    // Add date range filters
    if (params.due_date_from) queryParams.append("due_date_from", params.due_date_from);
    if (params.due_date_to) queryParams.append("due_date_to", params.due_date_to);

    // Add sorting
    if (params.sort_by) queryParams.append("sort_by", params.sort_by);
    if (params.sort_order) queryParams.append("sort_order", params.sort_order);

    // Add pagination
    if (params.page) queryParams.append("page", String(params.page));
    if (params.per_page) queryParams.append("per_page", String(params.per_page));

    const url = queryParams.toString() ? `${API_ENDPOINTS.TODOS_SEARCH}?${queryParams}` : API_ENDPOINTS.TODOS_SEARCH;
    const response = await this.get<TodoSearchResponse>(url);
    // dataプロパティが配列であることを保証
    if (response && typeof response === "object" && "data" in response) {
      return {
        ...response,
        data: Array.isArray(response.data) ? response.data : [],
      };
    }
    return {
      data: [],
      meta: {
        total: 0,
        current_page: 1,
        total_pages: 0,
        per_page: 20,
        filters_applied: {},
      },
    };
  }

  async createTodo(data: CreateTodoData, files?: File[]): Promise<Todo> {
    if (files && files.length > 0) {
      const formData = new FormData();
      formData.append("todo[title]", data.title);
      if (data.due_date) formData.append("todo[due_date]", data.due_date);
      if (data.priority) formData.append("todo[priority]", data.priority);
      if (data.status) formData.append("todo[status]", data.status);
      if (data.description) formData.append("todo[description]", data.description);
      if (data.category_id) formData.append("todo[category_id]", data.category_id.toString());
      if (data.tag_ids) {
        data.tag_ids.forEach((id) => formData.append("todo[tag_ids][]", id.toString()));
      }
      files.forEach((file) => formData.append("todo[files][]", file));

      return this.uploadFile<Todo>(API_ENDPOINTS.TODOS, formData);
    }
    return this.post<Todo>(API_ENDPOINTS.TODOS, { todo: data });
  }

  async updateTodo(id: number, data: UpdateTodoData, files?: File[]): Promise<Todo> {
    if (files && files.length > 0) {
      const formData = new FormData();
      if (data.title !== undefined) formData.append("todo[title]", data.title);
      if (data.completed !== undefined) formData.append("todo[completed]", data.completed.toString());
      if (data.due_date !== undefined) formData.append("todo[due_date]", data.due_date || "");
      if (data.priority !== undefined) formData.append("todo[priority]", data.priority);
      if (data.status !== undefined) formData.append("todo[status]", data.status);
      if (data.description !== undefined) formData.append("todo[description]", data.description || "");
      if (data.category_id !== undefined) formData.append("todo[category_id]", data.category_id?.toString() || "");
      if (data.tag_ids !== undefined) {
        data.tag_ids.forEach((id) => formData.append("todo[tag_ids][]", id.toString()));
      }
      files.forEach((file) => formData.append("todo[files][]", file));

      return this.uploadFile<Todo>(API_ENDPOINTS.TODO_BY_ID(id), formData, "PATCH");
    }
    return this.put<Todo>(API_ENDPOINTS.TODO_BY_ID(id), { todo: data });
  }

  async deleteTodo(id: number): Promise<void> {
    return this.delete<void>(API_ENDPOINTS.TODO_BY_ID(id));
  }

  async updateTodoOrder(todos: UpdateOrderData[]): Promise<void> {
    return this.patch<void>(API_ENDPOINTS.UPDATE_ORDER, { todos });
  }

  async updateTodoTags(id: number, tagIds: number[]): Promise<Todo> {
    return this.patch<Todo>(API_ENDPOINTS.UPDATE_TODO_TAGS(id), { tag_ids: tagIds });
  }

  // File operations
  async deleteTodoFile(todoId: number, fileId: string | number): Promise<Todo> {
    return this.delete<Todo>(API_ENDPOINTS.DELETE_TODO_FILE(todoId, fileId));
  }

  async downloadFile(url: string): Promise<Blob> {
    // For file downloads, we need to handle the response differently
    const token = localStorage.getItem("authToken");

    const response = await fetch(url, {
      headers: token ? { Authorization: `Bearer ${token}` } : {},
      credentials: "include",
    });

    if (!response.ok) {
      throw new ApiError(
        `HTTP ${response.status}: ${response.statusText}`,
        response.status,
      );
    }

    return response.blob();
  }
}

export const todoApiClient = new TodoApiClient();
export { ApiError };
