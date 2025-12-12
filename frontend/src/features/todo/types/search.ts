import type { Todo } from "./todo";

export interface SearchParams {
  q?: string;
  category_id?: number | -1;
  status?: string[];
  priority?: string[];
  tag_ids?: number[];
  tag_mode?: "any" | "all";
  due_date_from?: string;
  due_date_to?: string;
  sort_by?: "created_at" | "updated_at" | "due_date" | "title" | "priority" | "position";
  sort_order?: "asc" | "desc";
  page?: number;
  per_page?: number;
}

export interface SearchMeta {
  total: number;
  current_page: number;
  total_pages: number;
  per_page: number;
  search_query?: string;
  filters_applied: Record<string, unknown>;
}

export interface SearchSuggestion {
  type: "spelling" | "broader_search" | "clear_filters" | "reduce_filters" | "status_filter" | "tag_mode";
  message: string;
  current_filters?: string[];
}

export interface SearchResponse {
  todos: Todo[];
  meta: SearchMeta;
  suggestions?: SearchSuggestion[];
}
