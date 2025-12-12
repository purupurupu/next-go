"use client";

import { useAuth } from "@/contexts/auth-context";
import { Button } from "@/components/ui/button";
import Link from "next/link";
import { toast } from "sonner";

export function Navigation() {
  const { user, isAuthenticated, logout, isLoading } = useAuth();

  const handleLogout = async () => {
    try {
      await logout();
    } catch {
      toast.error("ログアウトに失敗しました");
    }
  };

  return (
    <nav className="bg-white border-b border-gray-200 px-4 py-3">
      <div className="max-w-7xl mx-auto flex justify-between items-center">
        <div className="flex items-center space-x-4">
          <Link href="/" className="text-xl font-bold text-gray-900">
            TODO App
          </Link>
          {isAuthenticated && (
            <>
              <Link href="/" className="text-sm text-gray-600 hover:text-gray-900">
                タスク
              </Link>
              <Link href="/categories" className="text-sm text-gray-600 hover:text-gray-900">
                カテゴリー
              </Link>
              <Link href="/tags" className="text-sm text-gray-600 hover:text-gray-900">
                タグ
              </Link>
              <Link href="/notes" className="text-sm text-gray-600 hover:text-gray-900">
                ノート
              </Link>
            </>
          )}
        </div>

        <div className="flex items-center space-x-4">
          {isAuthenticated
            ? (
                <>
                  {user && (
                    <span className="text-sm text-gray-600">
                      こんにちは、
                      {user.name}
                      さん
                    </span>
                  )}
                  <Button
                    variant="outline"
                    onClick={handleLogout}
                    disabled={isLoading}
                  >
                    {isLoading ? "ログアウト中..." : "ログアウト"}
                  </Button>
                </>
              )
            : (
                <Link href="/auth">
                  <Button variant="default">
                    ログイン
                  </Button>
                </Link>
              )}
        </div>
      </div>
    </nav>
  );
}
