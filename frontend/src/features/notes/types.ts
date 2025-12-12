export interface Note {
  id: number;
  title: string | null;
  body_md: string | null;
  pinned: boolean;
  archived_at: string | null;
  trashed_at: string | null;
  last_edited_at: string;
  created_at: string;
  updated_at: string;
  archived?: boolean;
  trashed?: boolean;
}

export interface NoteRevision {
  id: number;
  note_id: number;
  title: string | null;
  body_md: string | null;
  created_at: string;
}

export interface NotesMeta {
  total: number;
  current_page: number;
  total_pages: number;
  per_page: number;
  filters?: Record<string, unknown>;
}

export interface NotesListResponse {
  data: Note[];
  meta?: NotesMeta;
}

export interface NoteRevisionsResponse {
  data: NoteRevision[];
  meta?: NotesMeta;
}
