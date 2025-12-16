"use client";

import { useState, useEffect } from "react";
import { File, Image, FileText, FileSpreadsheet, Presentation, Loader2 } from "lucide-react";
import { cn } from "@/lib/utils";
import { API_BASE_URL } from "@/lib/constants";
import type { TodoFile } from "@/features/todo/types/todo";

interface FileThumbnailProps {
  file: TodoFile;
  todoId: number;
  size?: "small" | "medium" | "large";
  onClick?: () => void;
  className?: string;
}

const sizeClasses = {
  small: "h-12 w-12",
  medium: "h-20 w-20",
  large: "h-32 w-32",
};

const iconSizeClasses = {
  small: "h-5 w-5",
  medium: "h-8 w-8",
  large: "h-12 w-12",
};

export function FileThumbnail({
  file,
  todoId,
  size = "medium",
  onClick,
  className,
}: FileThumbnailProps) {
  const [thumbnailUrl, setThumbnailUrl] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(false);

  const isImage = file.file_type === "image";

  useEffect(() => {
    if (!isImage || !file.thumb_url) {
      return;
    }

    const loadThumbnail = async () => {
      setIsLoading(true);
      setError(false);

      try {
        const token = localStorage.getItem("authToken");
        const url = `${API_BASE_URL}${file.thumb_url}`;

        const response = await fetch(url, {
          headers: token ? { Authorization: `Bearer ${token}` } : {},
          credentials: "include",
        });

        if (!response.ok) {
          throw new Error("Failed to load thumbnail");
        }

        const blob = await response.blob();
        const objectUrl = URL.createObjectURL(blob);
        setThumbnailUrl(objectUrl);
      } catch {
        setError(true);
      } finally {
        setIsLoading(false);
      }
    };

    loadThumbnail();

    return () => {
      if (thumbnailUrl) {
        URL.revokeObjectURL(thumbnailUrl);
      }
    };
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isImage, file.thumb_url, todoId, file.id]);

  const getFileIcon = () => {
    const iconSize = iconSizeClasses[size];

    switch (file.file_type) {
      case "image":
        return <Image className={iconSize} />;
      case "document":
        if (file.content_type === "application/pdf") {
          return <FileText className={cn(iconSize, "text-red-500")} />;
        }
        if (file.content_type.includes("spreadsheet") || file.content_type.includes("excel")) {
          return <FileSpreadsheet className={cn(iconSize, "text-green-600")} />;
        }
        if (file.content_type.includes("presentation") || file.content_type.includes("powerpoint")) {
          return <Presentation className={cn(iconSize, "text-orange-500")} />;
        }
        if (file.content_type.includes("word")) {
          return <FileText className={cn(iconSize, "text-blue-600")} />;
        }
        return <FileText className={iconSize} />;
      default:
        return <File className={iconSize} />;
    }
  };

  const containerClasses = cn(
    sizeClasses[size],
    "rounded-lg overflow-hidden bg-muted flex items-center justify-center",
    onClick && "cursor-pointer hover:ring-2 hover:ring-primary/50 transition-all",
    className,
  );

  // Show loading state
  if (isLoading) {
    return (
      <div className={containerClasses}>
        <Loader2 className={cn(iconSizeClasses[size], "animate-spin text-muted-foreground")} />
      </div>
    );
  }

  // Show image thumbnail
  if (isImage && thumbnailUrl && !error) {
    return (
      <div className={containerClasses} onClick={onClick} role={onClick ? "button" : undefined}>
        <img
          src={thumbnailUrl}
          alt={file.original_name}
          className="h-full w-full object-cover"
        />
      </div>
    );
  }

  // Show file type icon
  return (
    <div className={containerClasses} onClick={onClick} role={onClick ? "button" : undefined}>
      {getFileIcon()}
    </div>
  );
}
