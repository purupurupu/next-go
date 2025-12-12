import { Navigation } from "@/components/navigation";
import { ProtectedRoute } from "@/components/protected-route";
import { CategoryManager } from "@/features/category/components/CategoryManager";

export default function CategoriesPage() {
  return (
    <div className="min-h-screen bg-background">
      <Navigation />
      <ProtectedRoute>
        <div className="container mx-auto px-4 py-8 max-w-4xl">
          <CategoryManager />
        </div>
      </ProtectedRoute>
    </div>
  );
}
