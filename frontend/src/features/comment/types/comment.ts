// Comment type definitions

export interface User {
  id: number;
  name: string;
  email: string;
}

export interface Comment {
  id: number;
  content: string;
  created_at: string;
  updated_at: string;
  editable: boolean;
  user: User;
}

export interface CreateCommentData {
  content: string;
}

export interface UpdateCommentData {
  content: string;
}
