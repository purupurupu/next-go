"use client";

import { useState } from "react";
import { File, Image, FileText, FileSpreadsheet, Archive, Download, Trash2, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { API_BASE_URL } from "@/lib/constants";
import type { TodoFile } from "@/features/todo/types/todo";
import { todoApiClient } from "@/features/todo/lib/api-client";
import { toast } from "sonner";
import { FileThumbnail } from "./FileThumbnail";
import { FilePreviewModal } from "./FilePreviewModal";

interface AttachmentListProps {
  todoId: number;
  files: TodoFile[];
  onDelete?: (fileId: number) => void;
  disabled?: boolean;
  compact?: boolean;
  showThumbnails?: boolean;
}

export function AttachmentList({
  todoId,
  files,
  onDelete,
  disabled = false,
  compact = false,
  showThumbnails = true,
}: AttachmentListProps) {
  const [downloadingIds, setDownloadingIds] = useState<Set<number>>(new Set());
  const [deletingIds, setDeletingIds] = useState<Set<number>>(new Set());
  const [previewOpen, setPreviewOpen] = useState(false);
  const [previewIndex, setPreviewIndex] = useState(0);

  // Get file name (supports both old and new format)
  const getFileName = (file: TodoFile): string => {
    return file.original_name || file.filename || "Unknown file";
  };

  // Get file size (supports both old and new format)
  const getFileSize = (file: TodoFile): number => {
    return file.file_size || file.byte_size || 0;
  };

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  const getFileIcon = (file: TodoFile) => {
    const fileType = file.file_type || "other";
    const contentType = file.content_type || "";

    if (fileType === "image" || contentType.startsWith("image/")) {
      return <Image className="h-4 w-4" aria-label="Image file" />;
    }
    if (contentType === "application/pdf" || contentType === "text/plain") {
      return <FileText className="h-4 w-4" />;
    }
    if (contentType.includes("spreadsheet") || contentType === "text/csv") {
      return <FileSpreadsheet className="h-4 w-4" />;
    }
    if (contentType.includes("zip") || contentType.includes("tar") || contentType.includes("gzip")) {
      return <Archive className="h-4 w-4" />;
    }
    return <File className="h-4 w-4" />;
  };

  const handleDownload = async (file: TodoFile) => {
    try {
      setDownloadingIds((prev) => new Set(prev).add(file.id));

      const token = localStorage.getItem("authToken");
      const url = `${API_BASE_URL}/api/v1/todos/${todoId}/files/${file.id}`;

      const response = await fetch(url, {
        headers: token ? { Authorization: `Bearer ${token}` } : {},
        credentials: "include",
      });

      if (!response.ok) {
        throw new Error("Download failed");
      }

      const blob = await response.blob();

      // Create a download link
      const objectUrl = window.URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = objectUrl;
      a.download = getFileName(file);
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(objectUrl);
      document.body.removeChild(a);

      toast.success(`${getFileName(file)} をダウンロードしました`);
    } catch {
      toast.error("ダウンロードに失敗しました");
    } finally {
      setDownloadingIds((prev) => {
        const newSet = new Set(prev);
        newSet.delete(file.id);
        return newSet;
      });
    }
  };

  const handleDelete = async (fileId: number) => {
    if (!onDelete) return;

    try {
      setDeletingIds((prev) => new Set(prev).add(fileId));
      await todoApiClient.deleteFile(todoId, fileId);
      onDelete(fileId);
      toast.success("ファイルを削除しました");
    } catch {
      toast.error("ファイルの削除に失敗しました");
    } finally {
      setDeletingIds((prev) => {
        const newSet = new Set(prev);
        newSet.delete(fileId);
        return newSet;
      });
    }
  };

  const handleThumbnailClick = (index: number) => {
    // Find the index in image files
    const imageFiles = files.filter((f) => f.file_type === "image");
    const imageIndex = imageFiles.findIndex((f) => f.id === files[index].id);

    if (imageIndex !== -1) {
      setPreviewIndex(imageIndex);
      setPreviewOpen(true);
    }
  };

  if (files.length === 0) {
    return null;
  }

  // Separate images and other files for grid display
  const imageFiles = files.filter((f) => f.file_type === "image");
  const otherFiles = files.filter((f) => f.file_type !== "image");

  if (compact) {
    return (
      <div className="flex flex-wrap gap-2">
        {files.map((file) => (
          <div
            key={file.id}
            className="flex items-center gap-1 rounded-md bg-muted px-2 py-1 text-xs"
          >
            {getFileIcon(file)}
            <span className="max-w-[100px] truncate">{getFileName(file)}</span>
            <span className="text-muted-foreground">
              ({formatFileSize(getFileSize(file))})
            </span>
          </div>
        ))}
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Image thumbnails grid */}
      {showThumbnails && imageFiles.length > 0 && (
        <div className="grid grid-cols-4 sm:grid-cols-6 md:grid-cols-8 gap-2">
          {imageFiles.map((file, index) => {
            const globalIndex = files.findIndex((f) => f.id === file.id);
            return (
              <div key={file.id} className="relative group">
                <FileThumbnail
                  file={file}
                  todoId={todoId}
                  size="small"
                  onClick={() => handleThumbnailClick(globalIndex)}
                  className="w-full aspect-square"
                />
                {onDelete && (
                  <Button
                    variant="destructive"
                    size="icon"
                    onClick={() => handleDelete(file.id)}
                    disabled={disabled || deletingIds.has(file.id)}
                    className="absolute -top-1 -right-1 h-5 w-5 rounded-full opacity-0 group-hover:opacity-100 transition-opacity"
                  >
                    {deletingIds.has(file.id) ? (
                      <Loader2 className="h-3 w-3 animate-spin" />
                    ) : (
                      <Trash2 className="h-3 w-3" />
                    )}
                  </Button>
                )}
              </div>
            );
          })}
        </div>
      )}

      {/* Other files list */}
      {otherFiles.length > 0 && (
        <div className="space-y-2">
          {otherFiles.map((file) => {
            const isDownloading = downloadingIds.has(file.id);
            const isDeleting = deletingIds.has(file.id);

            return (
              <div
                key={file.id}
                className={cn(
                  "flex items-center gap-3 rounded-lg border bg-card p-3",
                  (isDownloading || isDeleting) && "opacity-50",
                )}
              >
                <div className="flex items-center justify-center h-10 w-10 rounded-lg bg-muted">
                  {getFileIcon(file)}
                </div>

                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium truncate">{getFileName(file)}</p>
                  <p className="text-xs text-muted-foreground">
                    {formatFileSize(getFileSize(file))}
                  </p>
                </div>

                <div className="flex items-center gap-1">
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => handleDownload(file)}
                    disabled={disabled || isDownloading || isDeleting}
                    className="h-8 w-8"
                  >
                    {isDownloading ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                      <Download className="h-4 w-4" />
                    )}
                    <span className="sr-only">Download</span>
                  </Button>

                  {onDelete && (
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => handleDelete(file.id)}
                      disabled={disabled || isDeleting || isDownloading}
                      className="h-8 w-8 text-destructive hover:text-destructive"
                    >
                      {isDeleting ? (
                        <Loader2 className="h-4 w-4 animate-spin" />
                      ) : (
                        <Trash2 className="h-4 w-4" />
                      )}
                      <span className="sr-only">Delete</span>
                    </Button>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Image files list (when not showing thumbnails) */}
      {!showThumbnails && imageFiles.length > 0 && (
        <div className="space-y-2">
          {imageFiles.map((file) => {
            const isDownloading = downloadingIds.has(file.id);
            const isDeleting = deletingIds.has(file.id);

            return (
              <div
                key={file.id}
                className={cn(
                  "flex items-center gap-3 rounded-lg border bg-card p-3",
                  (isDownloading || isDeleting) && "opacity-50",
                )}
              >
                <div className="flex items-center justify-center h-10 w-10 rounded-lg bg-muted">
                  {getFileIcon(file)}
                </div>

                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium truncate">{getFileName(file)}</p>
                  <p className="text-xs text-muted-foreground">
                    {formatFileSize(getFileSize(file))}
                  </p>
                </div>

                <div className="flex items-center gap-1">
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => handleDownload(file)}
                    disabled={disabled || isDownloading || isDeleting}
                    className="h-8 w-8"
                  >
                    {isDownloading ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                      <Download className="h-4 w-4" />
                    )}
                    <span className="sr-only">Download</span>
                  </Button>

                  {onDelete && (
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => handleDelete(file.id)}
                      disabled={disabled || isDeleting || isDownloading}
                      className="h-8 w-8 text-destructive hover:text-destructive"
                    >
                      {isDeleting ? (
                        <Loader2 className="h-4 w-4 animate-spin" />
                      ) : (
                        <Trash2 className="h-4 w-4" />
                      )}
                      <span className="sr-only">Delete</span>
                    </Button>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Image preview modal */}
      <FilePreviewModal
        isOpen={previewOpen}
        onClose={() => setPreviewOpen(false)}
        files={files}
        initialIndex={previewIndex}
        todoId={todoId}
        onDownload={handleDownload}
      />
    </div>
  );
}
