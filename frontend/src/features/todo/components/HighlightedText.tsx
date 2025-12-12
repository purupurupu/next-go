import React from "react";

interface Highlight {
  start: number;
  end: number;
  matched_text: string;
}

interface HighlightedTextProps {
  text: string;
  highlights?: Highlight[];
  className?: string;
  highlightClassName?: string;
}

export function HighlightedText({
  text,
  highlights = [],
  className = "",
  highlightClassName = "bg-yellow-200 dark:bg-yellow-900/50 font-medium",
}: HighlightedTextProps) {
  if (!highlights.length) {
    return <span className={className}>{text}</span>;
  }

  // Sort highlights by start position
  const sortedHighlights = [...highlights].sort((a, b) => a.start - b.start);

  const parts: React.ReactNode[] = [];
  let lastIndex = 0;

  sortedHighlights.forEach((highlight, index) => {
    // Add text before highlight
    if (highlight.start > lastIndex) {
      parts.push(
        <span key={`text-${index}`}>
          {text.substring(lastIndex, highlight.start)}
        </span>,
      );
    }

    // Add highlighted text
    parts.push(
      <mark
        key={`highlight-${index}`}
        className={highlightClassName}
      >
        {text.substring(highlight.start, highlight.end)}
      </mark>,
    );

    lastIndex = highlight.end;
  });

  // Add remaining text
  if (lastIndex < text.length) {
    parts.push(
      <span key="text-end">
        {text.substring(lastIndex)}
      </span>,
    );
  }

  return <span className={className}>{parts}</span>;
}
