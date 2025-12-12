export const API_BASE_URL = process.env.NODE_ENV === "production"
  ? "https://your-production-api.com"
  : "http://localhost:3001";

export const API_ENDPOINTS = {
  TODOS: "/api/v1/todos",
  TODO_BY_ID: (id: number) => `/api/v1/todos/${id}`,
  TODOS_SEARCH: "/api/v1/todos/search",
  UPDATE_ORDER: "/api/v1/todos/update_order",
  UPDATE_TODO_TAGS: (id: number) => `/api/v1/todos/${id}/tags`,
  DELETE_TODO_FILE: (todoId: number, fileId: string | number) => `/api/v1/todos/${todoId}/files/${fileId}`,
  AUTH_LOGIN: "/auth/sign_in",
  AUTH_REGISTER: "/auth/sign_up",
  AUTH_LOGOUT: "/auth/sign_out",
  CATEGORIES: "/api/v1/categories",
  CATEGORY_BY_ID: (id: number) => `/api/v1/categories/${id}`,
  TAGS: "/api/v1/tags",
  TAG_BY_ID: (id: number) => `/api/v1/tags/${id}`,
  NOTES: "/api/v1/notes",
  NOTE_BY_ID: (id: number) => `/api/v1/notes/${id}`,
  NOTE_REVISIONS: (noteId: number) => `/api/v1/notes/${noteId}/revisions`,
  NOTE_REVISION_RESTORE: (noteId: number, revisionId: number) =>
    `/api/v1/notes/${noteId}/revisions/${revisionId}/restore`,
} as const;

export const TODO_FILTERS = {
  ALL: "all",
  ACTIVE: "active",
  COMPLETED: "completed",
} as const;

export const FILE_UPLOAD_CONSTANTS = {
  MAX_FILE_SIZE: 10 * 1024 * 1024, // 10MB
  MAX_TOTAL_SIZE: 50 * 1024 * 1024, // 50MB
  ALLOWED_FILE_TYPES: {
    // Images
    "image/jpeg": [".jpg", ".jpeg"],
    "image/png": [".png"],
    "image/gif": [".gif"],
    "image/webp": [".webp"],
    // Documents
    "application/pdf": [".pdf"],
    "application/msword": [".doc"],
    "application/vnd.openxmlformats-officedocument.wordprocessingml.document": [".docx"],
    "text/plain": [".txt"],
    // Spreadsheets
    "application/vnd.ms-excel": [".xls"],
    "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": [".xlsx"],
    "text/csv": [".csv"],
    // Archives
    "application/zip": [".zip"],
    "application/x-tar": [".tar"],
    "application/gzip": [".gz"],
  },
  ACCEPT_STRING: ".jpg,.jpeg,.png,.gif,.webp,.pdf,.doc,.docx,.txt,.xls,.xlsx,.csv,.zip,.tar,.gz",
} as const;
