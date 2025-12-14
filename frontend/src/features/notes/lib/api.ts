import { httpClient } from "@/lib/api-client";
import { Note, NoteRevisionsResponse, NotesListResponse } from "../types";

export interface NotesQuery {
  q?: string;
  pinned?: boolean;
  archived?: boolean;
  trashed?: boolean;
  page?: number;
  per_page?: number;
}

export async function fetchNotes(params: NotesQuery = {}): Promise<NotesListResponse> {
  const searchParams = new URLSearchParams();
  if (params.q) searchParams.set("q", params.q);
  if (typeof params.pinned === "boolean") searchParams.set("pinned", String(params.pinned));
  if (typeof params.archived === "boolean") searchParams.set("archived", String(params.archived));
  if (typeof params.trashed === "boolean") searchParams.set("trashed", String(params.trashed));
  if (params.page) searchParams.set("page", String(params.page));
  if (params.per_page) searchParams.set("per_page", String(params.per_page));

  const query = searchParams.toString();
  const endpoint = query ? `/api/v1/notes?${query}` : "/api/v1/notes";

  return httpClient.get<NotesListResponse>(endpoint);
}

export interface NotePayload {
  title?: string | null;
  body_md?: string | null;
  pinned?: boolean;
  archived?: boolean;
  trashed?: boolean;
}

export async function createNote(payload: NotePayload = {}): Promise<Note> {
  return httpClient.post<Note>("/api/v1/notes", { note: payload });
}

export async function updateNote(id: number, payload: NotePayload): Promise<Note> {
  return httpClient.patch<Note>(`/api/v1/notes/${id}`, { note: payload });
}

export async function deleteNote(id: number, force = false): Promise<void> {
  const endpoint = force ? `/api/v1/notes/${id}?force=true` : `/api/v1/notes/${id}`;
  await httpClient.delete<void>(endpoint);
}

export async function fetchRevisions(noteId: number): Promise<NoteRevisionsResponse> {
  return httpClient.get<NoteRevisionsResponse>(`/api/v1/notes/${noteId}/revisions`);
}

export async function restoreRevision(noteId: number, revisionId: number): Promise<Note> {
  return httpClient.post<Note>(`/api/v1/notes/${noteId}/revisions/${revisionId}/restore`);
}
