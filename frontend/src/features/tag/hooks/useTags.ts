"use client";

import { useState, useEffect, useCallback } from "react";
import { tagApiClient } from "../lib/api-client";
import type { Tag, CreateTagData, UpdateTagData } from "../types/tag";
import { toast } from "sonner";

export function useTags(fetchOnMount = true) {
  const [tags, setTags] = useState<Tag[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Fetch all tags
  const fetchTags = useCallback(async () => {
    try {
      setIsLoading(true);
      setError(null);
      const data = await tagApiClient.getTags();
      setTags(data);
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to fetch tags";
      setError(message);
      toast.error(message);
    } finally {
      setIsLoading(false);
    }
  }, []);

  // Create a new tag
  const createTag = useCallback(async (data: CreateTagData) => {
    try {
      const newTag = await tagApiClient.createTag(data);
      setTags((prev) => [...prev, newTag]);
      toast.success("Tag created successfully");
      return newTag;
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to create tag";
      toast.error(message);
      throw err;
    }
  }, []);

  // Update a tag
  const updateTag = useCallback(async (id: number, data: UpdateTagData) => {
    try {
      const updatedTag = await tagApiClient.updateTag(id, data);
      setTags((prev) =>
        prev.map((tag) => (tag.id === id ? updatedTag : tag)),
      );
      toast.success("Tag updated successfully");
      return updatedTag;
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to update tag";
      toast.error(message);
      throw err;
    }
  }, []);

  // Delete a tag
  const deleteTag = useCallback(async (id: number) => {
    try {
      await tagApiClient.deleteTag(id);
      setTags((prev) => prev.filter((tag) => tag.id !== id));
      toast.success("Tag deleted successfully");
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to delete tag";
      toast.error(message);
      throw err;
    }
  }, []);

  // Load tags on mount if requested
  useEffect(() => {
    if (fetchOnMount) {
      fetchTags();
    }
  }, [fetchOnMount]); // eslint-disable-line react-hooks/exhaustive-deps

  return {
    tags,
    isLoading,
    error,
    fetchTags,
    createTag,
    updateTag,
    deleteTag,
  };
}
