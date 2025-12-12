import { Navigation } from "@/components/navigation";
import { ProtectedRoute } from "@/components/protected-route";
import { TagManager } from "@/features/tag/components/TagManager";

export default function TagsPage() {
  return (
    <div className="min-h-screen bg-background">
      <Navigation />
      <ProtectedRoute>
        <div className="container mx-auto px-4 py-8 max-w-4xl">
          <TagManager />
        </div>
      </ProtectedRoute>
    </div>
  );
}
