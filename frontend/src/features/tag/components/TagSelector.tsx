"use client";

import { useState } from "react";
import { Check, ChevronsUpDown, Plus } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { TagBadge } from "./TagBadge";
import type { Tag } from "../types/tag";

interface TagSelectorProps {
  tags: Tag[];
  selectedTagIds: number[];
  onSelectionChange: (tagIds: number[]) => void;
  onCreateTag?: () => void;
  placeholder?: string;
  className?: string;
}

export function TagSelector({
  tags,
  selectedTagIds,
  onSelectionChange,
  onCreateTag,
  placeholder = "Select tags...",
  className,
}: TagSelectorProps) {
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState("");

  const selectedTags = tags?.filter((tag) => selectedTagIds.includes(tag.id)) || [];

  const toggleTag = (tagId: number) => {
    if (selectedTagIds.includes(tagId)) {
      onSelectionChange(selectedTagIds.filter((id) => id !== tagId));
    } else {
      onSelectionChange([...selectedTagIds, tagId]);
    }
  };

  const removeTag = (tagId: number) => {
    onSelectionChange(selectedTagIds.filter((id) => id !== tagId));
  };

  return (
    <div className={cn("space-y-2", className)}>
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            role="combobox"
            aria-expanded={open}
            className="w-full justify-between"
          >
            <span className="truncate">
              {selectedTags.length > 0
                ? `${selectedTags.length} tag${selectedTags.length > 1 ? "s" : ""} selected`
                : placeholder}
            </span>
            <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-full p-2">
          <div className="space-y-2">
            <Input
              placeholder="Search tags..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full"
            />
            <div className="max-h-[300px] overflow-y-auto">
              {(() => {
                const filteredTags = tags?.filter((tag) =>
                  tag.name.toLowerCase().includes(search.toLowerCase()),
                ) || [];

                if (filteredTags.length === 0) {
                  return (
                    <div className="p-4 text-center text-sm text-muted-foreground">
                      No tags found.
                      {onCreateTag && (
                        <Button
                          size="sm"
                          variant="ghost"
                          onClick={() => {
                            setOpen(false);
                            onCreateTag();
                          }}
                          className="mt-2 w-full"
                        >
                          <Plus className="mr-2 h-4 w-4" />
                          Create new tag
                        </Button>
                      )}
                    </div>
                  );
                }

                return (
                  <div className="space-y-1">
                    {filteredTags.map((tag) => (
                      <button
                        key={tag.id}
                        onClick={() => toggleTag(tag.id)}
                        className="flex items-center w-full rounded-md px-2 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground"
                      >
                        <Check
                          className={cn(
                            "mr-2 h-4 w-4",
                            selectedTagIds.includes(tag.id) ? "opacity-100" : "opacity-0",
                          )}
                        />
                        <TagBadge
                          name={tag.name}
                          color={tag.color}
                          className="mr-2"
                        />
                      </button>
                    ))}
                  </div>
                );
              })()}
            </div>
          </div>
        </PopoverContent>
      </Popover>

      {selectedTags.length > 0 && (
        <div className="flex flex-wrap gap-1">
          {selectedTags.map((tag) => (
            <TagBadge
              key={tag.id}
              name={tag.name}
              color={tag.color}
              onRemove={() => removeTag(tag.id)}
            />
          ))}
        </div>
      )}
    </div>
  );
}
