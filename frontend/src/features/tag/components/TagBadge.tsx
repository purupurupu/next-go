import { X } from "lucide-react";
import { cn } from "@/lib/utils";

interface TagBadgeProps {
  name: string;
  color: string;
  onRemove?: () => void;
  className?: string;
}

export function TagBadge({
  name,
  color,
  onRemove,
  className,
}: TagBadgeProps) {
  const textColor = getContrastColor(color);

  return (
    <div
      className={cn(
        "inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium",
        className,
      )}
      style={{
        backgroundColor: color,
        color: textColor,
      }}
    >
      <span>{name}</span>
      {onRemove && (
        <button
          onClick={onRemove}
          className="ml-0.5 hover:opacity-75 focus:outline-none"
          aria-label={`Remove ${name} tag`}
        >
          <X className="h-3 w-3" />
        </button>
      )}
    </div>
  );
}

// Helper function to determine if text should be light or dark based on background
function getContrastColor(hexColor: string): string {
  // Convert hex to RGB
  const hex = hexColor.replace("#", "");
  const r = parseInt(hex.substring(0, 2), 16);
  const g = parseInt(hex.substring(2, 4), 16);
  const b = parseInt(hex.substring(4, 6), 16);

  // Calculate luminance
  const luminance = (0.299 * r + 0.587 * g + 0.114 * b) / 255;

  // Return black or white based on luminance
  return luminance > 0.5 ? "#000000" : "#FFFFFF";
}
