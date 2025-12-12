import { Navigation } from "@/components/navigation";
import { ProtectedRoute } from "@/components/protected-route";
import { NotesWorkspace } from "@/features/notes/components/NotesWorkspace";

export default function NotesPage() {
  return (
    <div className="min-h-screen bg-background">
      <Navigation />
      <ProtectedRoute>
        <div className="container mx-auto px-4 py-8 max-w-6xl">
          <NotesWorkspace />
        </div>
      </ProtectedRoute>
    </div>
  );
}
