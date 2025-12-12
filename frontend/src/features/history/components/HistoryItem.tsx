"use client";

import { formatDistanceToNow } from "date-fns";
import { ja } from "date-fns/locale";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { TodoHistory } from "../types/history";
import { Clock, Edit3, CheckCircle, AlertCircle, TrendingUp } from "lucide-react";

interface HistoryItemProps {
  history: TodoHistory;
}

const actionIcons = {
  created: <Clock className="h-4 w-4 text-green-600" />,
  updated: <Edit3 className="h-4 w-4 text-blue-600" />,
  deleted: <AlertCircle className="h-4 w-4 text-red-600" />,
  status_changed: <CheckCircle className="h-4 w-4 text-purple-600" />,
  priority_changed: <TrendingUp className="h-4 w-4 text-orange-600" />,
};

export function HistoryItem({ history }: HistoryItemProps) {
  const formattedDate = formatDistanceToNow(new Date(history.created_at), {
    addSuffix: true,
    locale: ja,
  });

  return (
    <div className="flex gap-3 items-start">
      <div className="mt-1">
        {actionIcons[history.action] || <Clock className="h-4 w-4 text-gray-600" />}
      </div>

      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 mb-1">
          <Avatar className="h-6 w-6">
            <AvatarFallback className="text-xs">
              {history.user.name.charAt(0).toUpperCase()}
            </AvatarFallback>
          </Avatar>
          <span className="text-sm font-medium">{history.user.name}</span>
          <span className="text-xs text-muted-foreground">{formattedDate}</span>
        </div>
        <p className="text-sm text-muted-foreground">
          {history.human_readable_change}
        </p>
      </div>
    </div>
  );
}
