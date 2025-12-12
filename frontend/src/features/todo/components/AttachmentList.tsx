"use client";

import { useState } from "react";
import { File, Image, FileText, FileSpreadsheet, Archive, Download, Trash2, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import type { TodoFile } from "@/features/todo/types/todo";
import { todoApiClient } from "@/features/todo/lib/api-client";
import { toast } from "sonner";

interface AttachmentListProps {
  todoId: number;
  files: TodoFile[];
  onDelete?: (fileId: string | number) => void;
  disabled?: boolean;
  compact?: boolean;
}

export function AttachmentList({
  todoId,
  files,
  onDelete,
  disabled = false,
  compact = false,
}: AttachmentListProps) {
  const [downloadingIds, setDownloadingIds] = useState<Set<string | number>>(new Set());
  const [deletingIds, setDeletingIds] = useState<Set<string | number>>(new Set());

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  const getFileIcon = (fileType: string) => {
    if (fileType.startsWith("image/")) {
      return <Image className="h-4 w-4" aria-label="Image file" />;
    }
    if (fileType === "application/pdf" || fileType === "text/plain") {
      return <FileText className="h-4 w-4" />;
    }
    if (fileType.includes("spreadsheet") || fileType === "text/csv") {
      return <FileSpreadsheet className="h-4 w-4" />;
    }
    if (fileType.includes("zip") || fileType.includes("tar") || fileType.includes("gzip")) {
      return <Archive className="h-4 w-4" />;
    }
    return <File className="h-4 w-4" />;
  };

  const handleDownload = async (file: TodoFile) => {
    try {
      setDownloadingIds((prev) => new Set(prev).add(file.id));

      // Use the file URL directly for download
      const response = await fetch(file.url);
      const blob = await response.blob();

      // Create a download link
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = file.filename;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);

      toast.success(`Downloaded ${file.filename}`);
    } catch {
      toast.error("Failed to download file");
    } finally {
      setDownloadingIds((prev) => {
        const newSet = new Set(prev);
        newSet.delete(file.id);
        return newSet;
      });
    }
  };

  const handleDelete = async (fileId: string | number) => {
    if (!onDelete) return;

    try {
      setDeletingIds((prev) => new Set(prev).add(fileId));
      await todoApiClient.deleteTodoFile(todoId, fileId);
      onDelete(fileId);
      toast.success("File deleted");
    } catch {
      toast.error("Failed to delete file");
    } finally {
      setDeletingIds((prev) => {
        const newSet = new Set(prev);
        newSet.delete(fileId);
        return newSet;
      });
    }
  };

  if (files.length === 0) {
    return null;
  }

  if (compact) {
    return (
      <div className="flex flex-wrap gap-2">
        {files.map((file) => (
          <div
            key={file.id}
            className="flex items-center gap-1 rounded-md bg-muted px-2 py-1 text-xs"
          >
            {getFileIcon(file.content_type)}
            <span className="max-w-[100px] truncate">{file.filename}</span>
            <span className="text-muted-foreground">
              (
              {formatFileSize(file.byte_size)}
              )
            </span>
          </div>
        ))}
      </div>
    );
  }

  return (
    <div className="space-y-2">
      {files.map((file) => {
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
              {getFileIcon(file.content_type)}
            </div>

            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium truncate">{file.filename}</p>
              <p className="text-xs text-muted-foreground">
                {formatFileSize(file.byte_size)}
                {" "}
                •
                {file.url && "Available"}
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
                {isDownloading
                  ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    )
                  : (
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
                  {isDeleting
                    ? (
                        <Loader2 className="h-4 w-4 animate-spin" />
                      )
                    : (
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
  );
}
