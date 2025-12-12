import { HttpClient } from "@/lib/api-client";
import { API_ENDPOINTS } from "@/lib/constants";
import type { Tag, CreateTagData, UpdateTagData } from "../types/tag";

export class TagApiClient extends HttpClient {
  async getTags(): Promise<Tag[]> {
    const response = await this.get<Tag[]>(API_ENDPOINTS.TAGS);
    // 配列であることを保証
    return Array.isArray(response) ? response : [];
  }

  async getTag(id: number): Promise<Tag> {
    return this.get<Tag>(`${API_ENDPOINTS.TAGS}/${id}`);
  }

  async createTag(data: CreateTagData): Promise<Tag> {
    return this.post<Tag>(API_ENDPOINTS.TAGS, { tag: data });
  }

  async updateTag(id: number, data: UpdateTagData): Promise<Tag> {
    return this.patch<Tag>(`${API_ENDPOINTS.TAGS}/${id}`, { tag: data });
  }

  async deleteTag(id: number): Promise<void> {
    return this.delete(`${API_ENDPOINTS.TAGS}/${id}`);
  }
}

export const tagApiClient = new TagApiClient();
