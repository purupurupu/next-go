"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { Card } from "@/components/ui/card";
import { toast } from "sonner";
import { NoteDetailPanel } from "./NoteDetailPanel";
import { NoteListPanel } from "./NoteListPanel";
import {
  createNote,
  deleteNote,
  fetchNotes,
  fetchRevisions,
  restoreRevision,
  updateNote,
} from "../lib/api";
import { Note, NoteRevision } from "../types";

type ViewFilter = "active" | "archived" | "trashed";

const SAVE_DEBOUNCE_MS = 900;
const initialDraft = { title: "", body_md: "" };

const ensureNotesArray = (list: Note[] | null | undefined): Note[] => (Array.isArray(list) ? list : []);

export function NotesWorkspace() {
  const [notes, setNotes] = useState<Note[]>([]);
  const [selectedId, setSelectedId] = useState<number | null>(null);
  const [draft, setDraft] = useState<{ title: string; body_md: string }>(initialDraft);
  const [view, setView] = useState<ViewFilter>("active");
  const [search, setSearch] = useState("");
  const [revisions, setRevisions] = useState<NoteRevision[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [lastSavedAt, setLastSavedAt] = useState<string | null>(null);

  const safeNotes = useMemo(() => ensureNotesArray(notes), [notes]);

  const selectedNote = useMemo(
    () => safeNotes.find((note) => note.id === selectedId) || null,
    [safeNotes, selectedId],
  );

  const loadRevisions = useCallback(async (noteId: number) => {
    try {
      const response = await fetchRevisions(noteId);
      setRevisions(response.data);
    } catch {
      setRevisions([]);
    }
  }, []);

  const loadNotes = useCallback(async () => {
    setIsLoading(true);
    try {
      const response = await fetchNotes({
        q: search || undefined,
        archived: view === "archived" ? true : undefined,
        trashed: view === "trashed" ? true : undefined,
      });
      const next = ensureNotesArray(response.data);
      setNotes(next);

      if (selectedId && !next.some((n) => n.id === selectedId)) {
        setSelectedId(next[0]?.id ?? null);
      } else if (!selectedId) {
        setSelectedId(next[0]?.id ?? null);
      }
    } catch {
      toast.error("ノート一覧の取得に失敗しました");
    } finally {
      setIsLoading(false);
    }
  }, [search, selectedId, view]);

  useEffect(() => {
    loadNotes();
  }, [loadNotes]);

  useEffect(() => {
    if (selectedNote) {
      setDraft({
        title: selectedNote.title || "",
        body_md: selectedNote.body_md || "",
      });
      loadRevisions(selectedNote.id);
    } else {
      setDraft(initialDraft);
      setRevisions([]);
    }
  }, [selectedNote, loadRevisions]);

  // Autosave when draft changes
  useEffect(() => {
    if (!selectedNote) return;
    const hasChanges = draft.title !== (selectedNote.title || "") || draft.body_md !== (selectedNote.body_md || "");
    if (!hasChanges) return;

    const timer = setTimeout(() => {
      handleSaveDraft(selectedNote.id, draft);
    }, SAVE_DEBOUNCE_MS);

    return () => clearTimeout(timer);
  }, [draft, selectedNote]);

  async function handleSaveDraft(id: number, payload: { title: string; body_md: string }) {
    setIsSaving(true);
    try {
      const updated = await updateNote(id, payload);
      setNotes((prev) => ensureNotesArray(prev).map((note) => (note.id === id ? updated : note)));
      setLastSavedAt(new Date().toISOString());
    } catch {
      toast.error("保存に失敗しました");
    } finally {
      setIsSaving(false);
    }
  }

  async function handleCreateNote() {
    setIsSaving(true);
    try {
      const note = await createNote({ title: "新しいメモ", body_md: "" });
      setNotes((prev) => [note, ...ensureNotesArray(prev)]);
      setSelectedId(note.id);
      setLastSavedAt(note.updated_at);
      toast.success("新しいノートを作成しました");
    } catch {
      toast.error("ノートの作成に失敗しました");
    } finally {
      setIsSaving(false);
    }
  }

  async function handleArchiveToggle(note: Note) {
    try {
      const updated = await updateNote(note.id, { archived: !note.archived_at });

      setNotes((prev) => {
        const mapped = ensureNotesArray(prev).map((n) => (n.id === note.id ? updated : n));
        return mapped.filter((n) => {
          if (view === "active" && n.id === note.id && updated.archived_at) return false;
          if (view === "archived" && n.id === note.id && !updated.archived_at) return false;
          return true;
        });
      });

      if (view === "archived" && !updated.archived_at && selectedId === note.id) {
        setSelectedId(notes.find((n) => n.id !== note.id)?.id ?? null);
      }

      toast.success(updated.archived_at ? "アーカイブしました" : "アーカイブを解除しました");
    } catch {
      toast.error("アーカイブの更新に失敗しました");
    }
  }

  async function handleTrashToggle(note: Note) {
    try {
      const updated = await updateNote(note.id, { trashed: !note.trashed_at });
      setNotes((prev) => {
        const mapped = ensureNotesArray(prev).map((n) => (n.id === note.id ? updated : n));
        if (view === "trashed") {
          return mapped.filter((n) => !(n.id === note.id && !updated.trashed_at));
        }
        return mapped.filter((n) => !(n.id === note.id && updated.trashed_at));
      });
      if (updated.trashed_at && selectedId === note.id) {
        setSelectedId(null);
      }
      toast.success(updated.trashed_at ? "ゴミ箱に移動しました" : "ゴミ箱から復元しました");
    } catch {
      toast.error("ゴミ箱への移動に失敗しました");
    }
  }

  async function handleDeleteForever(note: Note) {
    try {
      await deleteNote(note.id, true);
      setNotes((prev) => ensureNotesArray(prev).filter((n) => n.id !== note.id));
      if (selectedId === note.id) {
        setSelectedId(null);
      }
      toast.success("ノートを完全に削除しました");
    } catch {
      toast.error("削除に失敗しました");
    }
  }

  async function handleRestoreRevision(revisionId: number) {
    if (!selectedNote) return;
    try {
      const restored = await restoreRevision(selectedNote.id, revisionId);
      setNotes((prev) => ensureNotesArray(prev).map((n) => (n.id === restored.id ? restored : n)));
      setDraft({
        title: restored.title || "",
        body_md: restored.body_md || "",
      });
      loadRevisions(restored.id);
      toast.success("リビジョンを復元しました");
    } catch {
      toast.error("リビジョンの復元に失敗しました");
    }
  }

  const previewText = useMemo(() => draft.body_md || "", [draft.body_md]);

  return (
    <div className="grid grid-cols-1 lg:grid-cols-[320px,1fr] gap-6">
      <NoteListPanel
        notes={safeNotes}
        selectedId={selectedId}
        onSelect={setSelectedId}
        search={search}
        onSearch={setSearch}
        view={view}
        onChangeView={setView}
        onCreate={handleCreateNote}
        isLoading={isLoading}
      />

      {!selectedNote
        ? (
            <Card className="p-4 min-h-[70vh] flex items-center justify-center text-muted-foreground">
              左のリストからノートを選択するか、新規作成してください。
            </Card>
          )
        : (
            <NoteDetailPanel
              note={selectedNote}
              draftTitle={draft.title}
              draftBody={previewText}
              onChangeTitle={(value) => setDraft((prev) => ({ ...prev, title: value }))}
              onChangeBody={(value) => setDraft((prev) => ({ ...prev, body_md: value }))}
              onArchiveToggle={() => handleArchiveToggle(selectedNote)}
              onTrashToggle={() => handleTrashToggle(selectedNote)}
              onDeleteForever={() => handleDeleteForever(selectedNote)}
              revisions={revisions}
              onRestoreRevision={handleRestoreRevision}
              isSaving={isSaving}
              lastSavedAt={lastSavedAt}
            />
          )}
    </div>
  );
}
