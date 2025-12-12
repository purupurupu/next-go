// Todo History type definitions

export interface TodoHistory {
  id: number;
  field_name: string;
  old_value: string | null;
  new_value: string | null;
  action: "created" | "updated" | "deleted" | "status_changed" | "priority_changed";
  created_at: string;
  human_readable_change: string;
  user: {
    id: number;
    name: string;
    email: string;
  };
}
