"use client";

import { useState, useEffect, useCallback } from "react";
import {
  Download,
  ZoomIn,
  ZoomOut,
  ChevronLeft,
  ChevronRight,
  Loader2,
  X,
} from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { API_BASE_URL } from "@/lib/constants";
import type { TodoFile } from "@/features/todo/types/todo";

interface FilePreviewModalProps {
  isOpen: boolean;
  onClose: () => void;
  files: TodoFile[];
  initialIndex: number;
  todoId: number;
  onDownload?: (file: TodoFile) => void;
}

const MIN_ZOOM = 0.5;
const MAX_ZOOM = 3;
const ZOOM_STEP = 0.25;

export function FilePreviewModal({
  isOpen,
  onClose,
  files,
  initialIndex,
  todoId,
  onDownload,
}: FilePreviewModalProps) {
  const [currentIndex, setCurrentIndex] = useState(initialIndex);
  const [zoom, setZoom] = useState(1);
  const [imageUrl, setImageUrl] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Filter to only image files
  const imageFiles = files.filter((f) => f.file_type === "image");
  const currentFile = imageFiles[currentIndex];

  const loadImage = useCallback(async () => {
    if (!currentFile) return;

    setIsLoading(true);
    setError(null);

    try {
      const token = localStorage.getItem("authToken");
      // Use medium size for preview if available, otherwise use original
      const url = currentFile.medium_url
        ? `${API_BASE_URL}${currentFile.medium_url}`
        : `${API_BASE_URL}/api/v1/todos/${todoId}/files/${currentFile.id}`;

      const response = await fetch(url, {
        headers: token ? { Authorization: `Bearer ${token}` } : {},
        credentials: "include",
      });

      if (!response.ok) {
        throw new Error("Failed to load image");
      }

      const blob = await response.blob();
      const objectUrl = URL.createObjectURL(blob);
      setImageUrl(objectUrl);
    } catch {
      setError("Failed to load image");
    } finally {
      setIsLoading(false);
    }
  }, [currentFile, todoId]);

  // Load image when current file changes
  useEffect(() => {
    if (isOpen && currentFile) {
      loadImage();
    }

    return () => {
      if (imageUrl) {
        URL.revokeObjectURL(imageUrl);
      }
    };
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isOpen, currentIndex, loadImage]);

  // Reset state when modal opens/closes
  useEffect(() => {
    if (isOpen) {
      setCurrentIndex(initialIndex);
      setZoom(1);
    } else {
      setImageUrl(null);
      setError(null);
    }
  }, [isOpen, initialIndex]);

  // Keyboard navigation
  useEffect(() => {
    if (!isOpen) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      switch (e.key) {
        case "ArrowLeft":
          e.preventDefault();
          goToPrevious();
          break;
        case "ArrowRight":
          e.preventDefault();
          goToNext();
          break;
        case "Escape":
          onClose();
          break;
        case "+":
        case "=":
          e.preventDefault();
          handleZoomIn();
          break;
        case "-":
          e.preventDefault();
          handleZoomOut();
          break;
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isOpen, currentIndex, imageFiles.length]);

  const goToPrevious = () => {
    setCurrentIndex((prev) => (prev > 0 ? prev - 1 : imageFiles.length - 1));
    setZoom(1);
  };

  const goToNext = () => {
    setCurrentIndex((prev) => (prev < imageFiles.length - 1 ? prev + 1 : 0));
    setZoom(1);
  };

  const handleZoomIn = () => {
    setZoom((prev) => Math.min(prev + ZOOM_STEP, MAX_ZOOM));
  };

  const handleZoomOut = () => {
    setZoom((prev) => Math.max(prev - ZOOM_STEP, MIN_ZOOM));
  };

  const handleDownload = async () => {
    if (!currentFile) return;

    if (onDownload) {
      onDownload(currentFile);
    } else {
      // Default download behavior
      try {
        const token = localStorage.getItem("authToken");
        const url = `${API_BASE_URL}/api/v1/todos/${todoId}/files/${currentFile.id}`;

        const response = await fetch(url, {
          headers: token ? { Authorization: `Bearer ${token}` } : {},
          credentials: "include",
        });

        if (!response.ok) throw new Error("Download failed");

        const blob = await response.blob();
        const objectUrl = URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = objectUrl;
        a.download = currentFile.original_name;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(objectUrl);
      } catch {
        // Error handling is done by the caller
      }
    }
  };

  if (imageFiles.length === 0) {
    return null;
  }

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent
        className="max-w-4xl h-[80vh] flex flex-col p-0 gap-0"
        showCloseButton={false}
      >
        {/* Header */}
        <DialogHeader className="flex-shrink-0 p-4 border-b">
          <div className="flex items-center justify-between">
            <DialogTitle className="truncate max-w-[60%]">
              {currentFile?.original_name}
            </DialogTitle>
            <div className="flex items-center gap-2">
              {/* Zoom controls */}
              <div className="flex items-center gap-1 mr-2">
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={handleZoomOut}
                  disabled={zoom <= MIN_ZOOM}
                  className="h-8 w-8"
                >
                  <ZoomOut className="h-4 w-4" />
                </Button>
                <span className="text-sm min-w-[3rem] text-center">
                  {Math.round(zoom * 100)}%
                </span>
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={handleZoomIn}
                  disabled={zoom >= MAX_ZOOM}
                  className="h-8 w-8"
                >
                  <ZoomIn className="h-4 w-4" />
                </Button>
              </div>

              {/* Download button */}
              <Button
                variant="ghost"
                size="icon"
                onClick={handleDownload}
                className="h-8 w-8"
              >
                <Download className="h-4 w-4" />
              </Button>

              {/* Close button */}
              <Button
                variant="ghost"
                size="icon"
                onClick={onClose}
                className="h-8 w-8"
              >
                <X className="h-4 w-4" />
              </Button>
            </div>
          </div>
        </DialogHeader>

        {/* Image container */}
        <div className="flex-1 relative overflow-auto bg-muted/50">
          {isLoading && (
            <div className="absolute inset-0 flex items-center justify-center">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          )}

          {error && !isLoading && (
            <div className="absolute inset-0 flex items-center justify-center">
              <p className="text-destructive">{error}</p>
            </div>
          )}

          {imageUrl && !isLoading && !error && (
            <div className="h-full flex items-center justify-center p-4">
              <img
                src={imageUrl}
                alt={currentFile?.original_name}
                className="max-w-full max-h-full object-contain transition-transform duration-200"
                style={{ transform: `scale(${zoom})` }}
              />
            </div>
          )}

          {/* Navigation arrows */}
          {imageFiles.length > 1 && (
            <>
              <Button
                variant="ghost"
                size="icon"
                onClick={goToPrevious}
                className={cn(
                  "absolute left-2 top-1/2 -translate-y-1/2 h-10 w-10 rounded-full bg-background/80 hover:bg-background shadow-md",
                )}
              >
                <ChevronLeft className="h-6 w-6" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                onClick={goToNext}
                className={cn(
                  "absolute right-2 top-1/2 -translate-y-1/2 h-10 w-10 rounded-full bg-background/80 hover:bg-background shadow-md",
                )}
              >
                <ChevronRight className="h-6 w-6" />
              </Button>
            </>
          )}
        </div>

        {/* Footer with pagination */}
        {imageFiles.length > 1 && (
          <div className="flex-shrink-0 p-2 border-t flex justify-center">
            <span className="text-sm text-muted-foreground">
              {currentIndex + 1} / {imageFiles.length}
            </span>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
