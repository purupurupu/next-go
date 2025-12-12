import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Separator } from "@/components/ui/separator";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import rehypeRaw from "rehype-raw";
import rehypeSanitize, { defaultSchema } from "rehype-sanitize";
import { Note, NoteRevision } from "../types";
import styles from "./NotesWorkspace.module.css";

const markdownSanitizeSchema = {
  ...defaultSchema,
  tagNames: [
    "a",
    "p",
    "br",
    "blockquote",
    "code",
    "pre",
    "em",
    "strong",
    "hr",
    "ul",
    "ol",
    "li",
    "h1",
    "h2",
    "h3",
    "h4",
    "h5",
    "h6",
    "table",
    "thead",
    "tbody",
    "tr",
    "th",
    "td",
  ],
  attributes: {
    ...(defaultSchema.attributes || {}),
    a: ["href", "title", "target", "rel"],
    code: ["className"],
    th: ["align"],
    td: ["align"],
  },
  clobberPrefix: "md-",
};

interface NoteDetailPanelProps {
  note: Note;
  draftTitle: string;
  draftBody: string;
  onChangeTitle: (value: string) => void;
  onChangeBody: (value: string) => void;
  onArchiveToggle: () => void;
  onTrashToggle: () => void;
  onDeleteForever: () => void;
  revisions: NoteRevision[];
  onRestoreRevision: (id: number) => void;
  isSaving: boolean;
  lastSavedAt: string | null;
}

export function NoteDetailPanel({
  note,
  draftTitle,
  draftBody,
  onChangeTitle,
  onChangeBody,
  onArchiveToggle,
  onTrashToggle,
  onDeleteForever,
  revisions,
  onRestoreRevision,
  isSaving,
  lastSavedAt,
}: NoteDetailPanelProps) {
  return (
    <div className="p-4 border rounded-lg space-y-4 min-h-[70vh]">
      <div className="flex items-center justify-between gap-2">
        <Input
          value={draftTitle}
          onChange={(e) => onChangeTitle(e.target.value)}
          placeholder="タイトル"
          className="font-semibold text-lg"
        />
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={onArchiveToggle}
          >
            {note.archived_at ? "アーカイブ解除" : "アーカイブ"}
          </Button>
          <Button
            variant={note.trashed_at ? "default" : "destructive"}
            size="sm"
            onClick={onTrashToggle}
          >
            {note.trashed_at ? "復元" : "ゴミ箱へ"}
          </Button>
          {note.trashed_at && (
            <Button
              variant="destructive"
              size="sm"
              onClick={onDeleteForever}
            >
              完全削除
            </Button>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <Textarea
          value={draftBody}
          onChange={(e) => onChangeBody(e.target.value)}
          placeholder="Markdownでメモを書いてください"
          className="min-h-[260px] font-mono text-sm"
        />
        <div className="space-y-3">
          <div className="space-y-2">
            <p className="text-sm font-semibold">プレビュー</p>
            <div className={`rounded border bg-muted/30 p-3 min-h-[260px] ${styles.markdownPreview}`}>
              {draftBody.trim()
                ? (
                    <ReactMarkdown
                      remarkPlugins={[remarkGfm]}
                      rehypePlugins={[rehypeRaw, [rehypeSanitize, markdownSanitizeSchema]]}
                    >
                      {draftBody}
                    </ReactMarkdown>
                  )
                : (
                    <p className="text-muted-foreground text-sm">何も入力されていません</p>
                  )}
            </div>
          </div>
          <Separator />
          <div className="space-y-2">
            <p className="text-sm font-semibold">リビジョン</p>
            <div className="space-y-2 max-h-[200px] overflow-y-auto pr-1">
              {revisions.length === 0 && (
                <p className="text-xs text-muted-foreground">リビジョンがありません</p>
              )}
              {revisions.map((rev) => (
                <div
                  key={rev.id}
                  className="rounded border p-2 text-xs flex items-center justify-between gap-2"
                >
                  <div>
                    <p className="font-semibold line-clamp-1">{rev.title || "無題"}</p>
                    <p className="text-muted-foreground">
                      {new Date(rev.created_at).toLocaleString()}
                    </p>
                  </div>
                  <Button
                    size="sm"
                    variant="outline"
                    onClick={() => onRestoreRevision(rev.id)}
                  >
                    復元
                  </Button>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>

      <div className="flex items-center justify-between text-xs text-muted-foreground">
        <span>
          {isSaving
            ? "保存中..."
            : lastSavedAt
              ? `保存済み: ${new Date(lastSavedAt).toLocaleTimeString()}`
              : "自動保存が有効です"}
        </span>
        <span>
          リビジョン:
          {" "}
          {revisions.length}
          件
        </span>
      </div>
    </div>
  );
}
