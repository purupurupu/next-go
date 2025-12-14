export const API_BASE_URL = process.env.NODE_ENV === "production"
  ? "https://your-production-api.com"
  : "http://localhost:3001";

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
