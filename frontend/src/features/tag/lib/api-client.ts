import { ApiClient } from "@/lib/api-client";
import type { Tag, CreateTagData, UpdateTagData } from "../types/tag";

export class TagApiClient extends ApiClient {
  async getTags(): Promise<Tag[]> {
    const response = await this.get<Tag[]>("/tags");
    // 配列であることを保証
    return Array.isArray(response) ? response : [];
  }

  async getTag(id: number): Promise<Tag> {
    return this.get<Tag>(`/tags/${id}`);
  }

  async createTag(data: CreateTagData): Promise<Tag> {
    return this.post<Tag>("/tags", data);
  }

  async updateTag(id: number, data: UpdateTagData): Promise<Tag> {
    return this.patch<Tag>(`/tags/${id}`, data);
  }

  async deleteTag(id: number): Promise<void> {
    return this.delete(`/tags/${id}`);
  }
}

export const tagApiClient = new TagApiClient();
