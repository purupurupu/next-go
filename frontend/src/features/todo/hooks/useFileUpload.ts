import { useState, useCallback } from "react";
import { toast } from "sonner";
import { todoApiClient, ApiError } from "@/features/todo/lib/api-client";
import type { TodoFile } from "../types/todo";

// File size limit: 10MB
const MAX_FILE_SIZE = 10 * 1024 * 1024;

// Allowed MIME types (matches backend)
const ALLOWED_MIME_TYPES = [
  // Images
  "image/jpeg",
  "image/png",
  "image/gif",
  "image/webp",
  // Documents
  "application/pdf",
  "text/plain",
  // MS Office
  "application/msword",
  "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
  "application/vnd.ms-excel",
  "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
  "application/vnd.ms-powerpoint",
  "application/vnd.openxmlformats-officedocument.presentationml.presentation",
];

export interface UploadingFile {
  id: string;
  file: File;
  progress: number;
  status: "pending" | "uploading" | "success" | "error";
  error?: string;
  uploadedFile?: TodoFile;
}

export interface FileValidationError {
  fileName: string;
  message: string;
}

export interface UseFileUploadParams {
  todoId: number;
  onUploadComplete?: (file: TodoFile) => void;
  onAllUploadsComplete?: (files: TodoFile[]) => void;
}

export interface UseFileUploadReturn {
  uploadingFiles: UploadingFile[];
  validationErrors: FileValidationError[];
  isUploading: boolean;
  addFiles: (files: FileList | File[]) => void;
  removeFile: (fileId: string) => void;
  uploadFiles: () => Promise<TodoFile[]>;
  uploadSingleFile: (file: File) => Promise<TodoFile | null>;
  clearFiles: () => void;
  clearErrors: () => void;
}

function generateId(): string {
  return `${Date.now()}-${Math.random().toString(36).substring(2, 9)}`;
}

function validateFile(file: File): FileValidationError | null {
  if (file.size > MAX_FILE_SIZE) {
    return {
      fileName: file.name,
      message: `ファイルサイズが大きすぎます (最大 ${MAX_FILE_SIZE / 1024 / 1024}MB)`,
    };
  }

  if (!ALLOWED_MIME_TYPES.includes(file.type)) {
    return {
      fileName: file.name,
      message: "許可されていないファイル形式です",
    };
  }

  return null;
}

export function useFileUpload({
  todoId,
  onUploadComplete,
  onAllUploadsComplete,
}: UseFileUploadParams): UseFileUploadReturn {
  const [uploadingFiles, setUploadingFiles] = useState<UploadingFile[]>([]);
  const [validationErrors, setValidationErrors] = useState<FileValidationError[]>([]);

  const isUploading = uploadingFiles.some((f) => f.status === "uploading");

  const addFiles = useCallback((files: FileList | File[]) => {
    const fileArray = Array.from(files);
    const newErrors: FileValidationError[] = [];
    const validFiles: UploadingFile[] = [];

    for (const file of fileArray) {
      const error = validateFile(file);
      if (error) {
        newErrors.push(error);
      } else {
        validFiles.push({
          id: generateId(),
          file,
          progress: 0,
          status: "pending",
        });
      }
    }

    if (newErrors.length > 0) {
      setValidationErrors((prev) => [...prev, ...newErrors]);
      newErrors.forEach((error) => {
        toast.error(`${error.fileName}: ${error.message}`);
      });
    }

    if (validFiles.length > 0) {
      setUploadingFiles((prev) => [...prev, ...validFiles]);
    }
  }, []);

  const removeFile = useCallback((fileId: string) => {
    setUploadingFiles((prev) => prev.filter((f) => f.id !== fileId));
  }, []);

  const uploadSingleFile = useCallback(async (file: File): Promise<TodoFile | null> => {
    const validationError = validateFile(file);
    if (validationError) {
      toast.error(`${validationError.fileName}: ${validationError.message}`);
      return null;
    }

    try {
      const uploadedFile = await todoApiClient.uploadTodoFile(todoId, file);
      onUploadComplete?.(uploadedFile);
      toast.success(`${file.name} をアップロードしました`);
      return uploadedFile;
    } catch (error) {
      const errorMessage = error instanceof ApiError
        ? error.message
        : "ファイルのアップロードに失敗しました";
      toast.error(errorMessage);
      return null;
    }
  }, [todoId, onUploadComplete]);

  const uploadFiles = useCallback(async (): Promise<TodoFile[]> => {
    const pendingFiles = uploadingFiles.filter((f) => f.status === "pending");
    if (pendingFiles.length === 0) {
      return [];
    }

    const uploadedFiles: TodoFile[] = [];

    for (const uploadingFile of pendingFiles) {
      // Set uploading status
      setUploadingFiles((prev) =>
        prev.map((f) =>
          f.id === uploadingFile.id
            ? { ...f, status: "uploading" as const, progress: 0 }
            : f
        )
      );

      try {
        const uploadedFile = await todoApiClient.uploadTodoFile(todoId, uploadingFile.file);

        // Set success status
        setUploadingFiles((prev) =>
          prev.map((f) =>
            f.id === uploadingFile.id
              ? { ...f, status: "success" as const, progress: 100, uploadedFile }
              : f
          )
        );

        uploadedFiles.push(uploadedFile);
        onUploadComplete?.(uploadedFile);
        toast.success(`${uploadingFile.file.name} をアップロードしました`);
      } catch (error) {
        const errorMessage = error instanceof ApiError
          ? error.message
          : "ファイルのアップロードに失敗しました";

        // Set error status
        setUploadingFiles((prev) =>
          prev.map((f) =>
            f.id === uploadingFile.id
              ? { ...f, status: "error" as const, error: errorMessage }
              : f
          )
        );

        toast.error(`${uploadingFile.file.name}: ${errorMessage}`);
      }
    }

    if (uploadedFiles.length > 0) {
      onAllUploadsComplete?.(uploadedFiles);
    }

    return uploadedFiles;
  }, [todoId, uploadingFiles, onUploadComplete, onAllUploadsComplete]);

  const clearFiles = useCallback(() => {
    setUploadingFiles([]);
  }, []);

  const clearErrors = useCallback(() => {
    setValidationErrors([]);
  }, []);

  return {
    uploadingFiles,
    validationErrors,
    isUploading,
    addFiles,
    removeFile,
    uploadFiles,
    uploadSingleFile,
    clearFiles,
    clearErrors,
  };
}
