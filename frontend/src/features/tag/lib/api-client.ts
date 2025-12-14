import { HttpClient } from "@/lib/api-client";
import type { Tag, CreateTagData, UpdateTagData } from "../types/tag";

export class TagApiClient extends HttpClient {
  async getTags(): Promise<Tag[]> {
    const response = await this.get<Tag[]>("/api/v1/tags");
    // 配列であることを保証
    return Array.isArray(response) ? response : [];
  }

  async getTag(id: number): Promise<Tag> {
    return this.get<Tag>(`/api/v1/tags/${id}`);
  }

  async createTag(data: CreateTagData): Promise<Tag> {
    return this.post<Tag>("/api/v1/tags", { tag: data });
  }

  async updateTag(id: number, data: UpdateTagData): Promise<Tag> {
    return this.patch<Tag>(`/api/v1/tags/${id}`, { tag: data });
  }

  async deleteTag(id: number): Promise<void> {
    return this.delete(`/api/v1/tags/${id}`);
  }
}

export const tagApiClient = new TagApiClient();
