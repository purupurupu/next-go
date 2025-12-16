import { SearchBar } from "@/features/todo/components/SearchBar";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Note } from "../types";

type ViewFilter = "active" | "archived" | "trashed";

interface NoteListPanelProps {
  notes: Note[];
  selectedId: number | null;
  onSelect: (id: number | null) => void;
  search: string;
  onSearch: (value: string) => void;
  view: ViewFilter;
  onChangeView: (v: ViewFilter) => void;
  onCreate: () => void;
  isLoading: boolean;
}

export function NoteListPanel({
  notes,
  selectedId,
  onSelect,
  search,
  onSearch,
  view,
  onChangeView,
  onCreate,
  isLoading,
}: NoteListPanelProps) {
  return (
    <div className="p-4 border rounded-lg space-y-4">
      <div className="flex items-center justify-between gap-2">
        <div>
          <p className="text-sm font-semibold">ノート</p>
          <p className="text-xs text-muted-foreground">検索とフィルター</p>
        </div>
        <Button size="sm" onClick={onCreate}>
          新規作成
        </Button>
      </div>
      <SearchBar
        value={search}
        onChange={onSearch}
        placeholder="ノートを検索..."
        debounceDelay={400}
      />
      <div className="flex flex-wrap gap-2">
        <FilterButton active={view === "active"} onClick={() => onChangeView("active")}>
          アクティブ
        </FilterButton>
        <FilterButton active={view === "archived"} onClick={() => onChangeView("archived")}>
          アーカイブ
        </FilterButton>
        <FilterButton active={view === "trashed"} onClick={() => onChangeView("trashed")}>
          ゴミ箱
        </FilterButton>
      </div>
      <Separator />
      <div className="space-y-2 max-h-[65vh] overflow-y-auto pr-1">
        {isLoading && <p className="text-sm text-muted-foreground">読み込み中...</p>}
        {!isLoading && notes.length === 0 && (
          <p className="text-sm text-muted-foreground">ノートがありません</p>
        )}
        {notes.map((note) => (
          <button
            key={`note-${note.id}`}
            type="button"
            onClick={() => onSelect(note.id)}
            className={`w-full text-left rounded-md border px-3 py-2 transition ${
              selectedId === note.id ? "border-primary bg-primary/5" : "border-border hover:bg-muted"
            }`}
          >
            <div className="flex items-center justify-between gap-2">
              <span className="font-semibold text-sm line-clamp-1">
                {note.title || "無題のノート"}
              </span>
              <div className="flex gap-1">
                {note.archived_at && <Badge variant="outline" className="text-[10px]">Arch</Badge>}
                {note.trashed_at && <Badge variant="destructive" className="text-[10px]">Trash</Badge>}
              </div>
            </div>
            <p className="text-xs text-muted-foreground line-clamp-2">
              {(note.body_md || "").slice(0, 120) || "本文なし"}
            </p>
            <p className="text-[11px] text-muted-foreground">
              更新:
              {" "}
              {new Date(note.updated_at).toLocaleString()}
            </p>
          </button>
        ))}
      </div>
    </div>
  );
}

function FilterButton({
  active,
  onClick,
  children,
}: {
  active: boolean;
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <Button
      variant={active ? "default" : "outline"}
      size="sm"
      onClick={onClick}
    >
      {children}
    </Button>
  );
}
