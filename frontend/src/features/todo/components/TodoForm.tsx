"use client";

import { useState, useEffect } from "react";
import { format } from "date-fns";
import { ja } from "date-fns/locale";
import { Calendar as CalendarIcon } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Calendar } from "@/components/ui/calendar";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";

import type { CreateTodoData, Todo, TodoPriority, TodoStatus } from "@/features/todo/types/todo";
import { useCategories } from "@/features/category/hooks/useCategories";
import { useTags } from "@/features/tag/hooks/useTags";
import { TagSelector } from "@/features/tag/components/TagSelector";
import { FileUpload } from "@/features/todo/components/FileUpload";
import { AttachmentList } from "@/features/todo/components/AttachmentList";
import { CommentList } from "@/features/comment/components/CommentList";
import { HistoryList } from "@/features/history/components/HistoryList";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { cn } from "@/lib/utils";

interface TodoFormProps {
  mode: "create" | "edit";
  todo?: Todo;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (data: CreateTodoData, files?: File[]) => Promise<void>;
  onFileDelete?: (fileId: string | number) => void;
}

export function TodoForm({ mode, todo, open, onOpenChange, onSubmit, onFileDelete }: TodoFormProps) {
  const { categories, fetchCategories } = useCategories(false);
  const { tags, fetchTags } = useTags(false);
  const [title, setTitle] = useState(todo?.title || "");
  const [dueDate, setDueDate] = useState<Date | undefined>(
    todo?.due_date ? new Date(todo.due_date) : undefined,
  );
  const [priority, setPriority] = useState<TodoPriority>(todo?.priority || "medium");
  const [status, setStatus] = useState<TodoStatus>(todo?.status || "pending");
  const [description, setDescription] = useState(todo?.description || "");
  const [categoryId, setCategoryId] = useState<number | undefined>(todo?.category?.id);
  const [selectedTagIds, setSelectedTagIds] = useState<number[]>(
    todo?.tags?.map((tag) => tag.id) || [],
  );
  const [selectedFiles, setSelectedFiles] = useState<File[]>([]);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [showCalendar, setShowCalendar] = useState(false);

  // Fetch categories and tags when form opens
  useEffect(() => {
    if (open) {
      fetchCategories();
      fetchTags();
    }
  }, [open, fetchCategories, fetchTags]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) return;

    setIsSubmitting(true);
    try {
      const data = {
        title: title.trim(),
        due_date: dueDate ? format(dueDate, "yyyy-MM-dd") : null,
        priority,
        status,
        description: description.trim() || null,
        category_id: categoryId || null,
        tag_ids: selectedTagIds,
      };

      // Submit with files if any are selected
      if (selectedFiles.length > 0) {
        await onSubmit(data, selectedFiles);
      } else {
        await onSubmit(data);
      }

      // Reset form for create mode
      if (mode === "create") {
        setTitle("");
        setDueDate(undefined);
        setPriority("medium");
        setStatus("pending");
        setDescription("");
        setCategoryId(undefined);
        setSelectedTagIds([]);
        setSelectedFiles([]);
      }

      onOpenChange(false);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleOpenChange = (newOpen: boolean) => {
    if (!newOpen && mode === "create") {
      setTitle("");
      setDueDate(undefined);
      setPriority("medium");
      setStatus("pending");
      setDescription("");
      setCategoryId(undefined);
      setSelectedTagIds([]);
      setSelectedFiles([]);
    }
    onOpenChange(newOpen);
  };

  const clearDueDate = () => {
    setDueDate(undefined);
    setShowCalendar(false);
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent
        className={cn(
          "sm:max-w-md flex flex-col",
          mode === "edit" ? "max-h-[95vh]" : "max-h-[85vh]",
        )}
      >
        <DialogHeader className="flex-shrink-0">
          <DialogTitle>
            {mode === "create" ? "タスクを追加" : "タスクを編集"}
          </DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="flex flex-col flex-1 min-h-0 overflow-hidden">
          <div className="flex-1 overflow-y-auto px-1 py-4 space-y-4">
            <div className="space-y-2">
              <label htmlFor="title" className="text-sm font-medium">
                タスク名
              </label>
              <Input
                id="title"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="タスクを入力してください"
                required
              />
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium">優先度</label>
              <Select value={priority} onValueChange={(value: TodoPriority) => setPriority(value)}>
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="優先度を選択" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="low">低</SelectItem>
                  <SelectItem value="medium">中</SelectItem>
                  <SelectItem value="high">高</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium">ステータス</label>
              <Select value={status} onValueChange={(value: TodoStatus) => setStatus(value)}>
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="ステータスを選択" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="pending">未着手</SelectItem>
                  <SelectItem value="in_progress">進行中</SelectItem>
                  <SelectItem value="completed">完了</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium">カテゴリー（任意）</label>
              <Select
                value={categoryId?.toString() || "none"}
                onValueChange={(value) => setCategoryId(value === "none" ? undefined : parseInt(value))}
              >
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="カテゴリーを選択">
                    {categoryId
                      ? (
                          <div className="flex items-center gap-2">
                            <div
                              className="h-3 w-3 rounded"
                              style={{ backgroundColor: categories.find((c) => c.id === categoryId)?.color }}
                            />
                            {categories.find((c) => c.id === categoryId)?.name}
                          </div>
                        )
                      : (
                          "カテゴリーなし"
                        )}
                  </SelectValue>
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">カテゴリーなし</SelectItem>
                  {categories.map((category) => (
                    <SelectItem key={category.id} value={category.id.toString()}>
                      <div className="flex items-center gap-2">
                        <div
                          className="h-3 w-3 rounded"
                          style={{ backgroundColor: category.color }}
                        />
                        {category.name}
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium">タグ（任意）</label>
              <TagSelector
                tags={tags}
                selectedTagIds={selectedTagIds}
                onSelectionChange={setSelectedTagIds}
                placeholder="タグを選択..."
              />
            </div>

            <div className="space-y-2">
              <label htmlFor="description" className="text-sm font-medium">
                説明（任意）
              </label>
              <Textarea
                id="description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="タスクの詳細説明を入力してください"
                rows={3}
              />
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium">期限日</label>
              <div className="space-y-2">
                {dueDate
                  ? (
                      <div className="flex items-center gap-2">
                        <Badge variant="outline" className="flex items-center gap-1">
                          <CalendarIcon className="h-3 w-3" />
                          {format(dueDate, "yyyy年M月d日", { locale: ja })}
                        </Badge>
                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          onClick={clearDueDate}
                        >
                          クリア
                        </Button>
                      </div>
                    )
                  : null}

                <Select
                  value={showCalendar ? "custom" : ""}
                  onValueChange={(value) => {
                    if (value === "today") {
                      setDueDate(new Date());
                    } else if (value === "tomorrow") {
                      const tomorrow = new Date();
                      tomorrow.setDate(tomorrow.getDate() + 1);
                      setDueDate(tomorrow);
                    } else if (value === "week") {
                      const nextWeek = new Date();
                      nextWeek.setDate(nextWeek.getDate() + 7);
                      setDueDate(nextWeek);
                    } else if (value === "custom") {
                      setShowCalendar(true);
                    }
                  }}
                >
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="期限日を設定" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="today">今日</SelectItem>
                    <SelectItem value="tomorrow">明日</SelectItem>
                    <SelectItem value="week">1週間後</SelectItem>
                    <SelectItem value="custom">日付を選択</SelectItem>
                  </SelectContent>
                </Select>

                {showCalendar && (
                  <div className="border rounded-md p-3">
                    <Calendar
                      mode="single"
                      selected={dueDate}
                      onSelect={(date) => {
                        setDueDate(date);
                        setShowCalendar(false);
                      }}
                      disabled={(date) => date < new Date()}
                      className="mx-auto"
                    />
                  </div>
                )}
              </div>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium">添付ファイル</label>
              {mode === "edit" && todo && todo.files && todo.files.length > 0 && (
                <AttachmentList
                  todoId={todo.id}
                  files={todo.files}
                  onDelete={onFileDelete}
                  disabled={isSubmitting}
                />
              )}
              <FileUpload
                onFileSelect={setSelectedFiles}
                existingFiles={selectedFiles}
                disabled={isSubmitting}
              />
            </div>

            {/* 編集モードでのみコメントと履歴を表示 */}
            {mode === "edit" && todo && (
              <div className="mt-6">
                <Tabs defaultValue="comments" className="w-full">
                  <TabsList className="grid w-full grid-cols-2">
                    <TabsTrigger value="comments">コメント</TabsTrigger>
                    <TabsTrigger value="history">変更履歴</TabsTrigger>
                  </TabsList>
                  <TabsContent value="comments" className="mt-4">
                    <CommentList todoId={todo.id} />
                  </TabsContent>
                  <TabsContent value="history" className="mt-4">
                    <HistoryList todoId={todo.id} />
                  </TabsContent>
                </Tabs>
              </div>
            )}
          </div>

          <DialogFooter className="flex-shrink-0 pt-4 border-t">
            <Button
              type="button"
              variant="outline"
              onClick={() => handleOpenChange(false)}
              disabled={isSubmitting}
            >
              キャンセル
            </Button>
            <Button type="submit" disabled={isSubmitting || !title.trim()}>
              {isSubmitting ? "処理中..." : mode === "create" ? "追加" : "更新"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
