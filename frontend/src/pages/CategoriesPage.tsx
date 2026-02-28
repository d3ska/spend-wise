import { useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { Plus, Pencil, Tags, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Skeleton } from "@/components/ui/skeleton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog";
import {
  useCategories,
  useCreateCategory,
  useUpdateCategory,
  useDeleteCategory,
} from "@/api/categories";
import { useWorkspaceContext } from "@/hooks/useWorkspace";
import { getErrorMessage } from "@/lib/errors";
import { getCategoryDisplayName, isUncategorized } from "@/lib/categoryUtils";
import type { Category } from "@/types";

const CATEGORY_ICONS = [
  "\u{1F6D2}", // Groceries
  "\u{1F355}", // Dining
  "\u{1F697}", // Transport
  "\u{1F3E0}", // Housing
  "\u{1F4A1}", // Utilities
  "\u{1F3E5}", // Health
  "\u{1F48A}", // Pharmacy
  "\u{1F455}", // Clothing
  "\u{1F393}", // Education
  "\u{1F3AE}", // Entertainment
  "\u{1F4BB}", // Tech
  "\u{1F4F1}", // Phone
  "\u2708\uFE0F", // Travel
  "\u{1F381}", // Gifts
  "\u{1F487}", // Personal Care
  "\u{1F43E}", // Pets
  "\u{1F476}", // Kids
  "\u{1F3CB}\uFE0F", // Fitness
  "\u{1F4DA}", // Books
  "\u2615",    // Coffee
  "\u{1F3AC}", // Movies
  "\u{1F37A}", // Drinks
  "\u{1F6E0}\uFE0F", // Repairs
  "\u{1F4E6}", // Subscriptions
  "\u{1F3E6}", // Banking
  "\u{1F4B0}", // Savings
  "\u{1F3B5}", // Music
  "\u{1F9F9}", // Household
  "\u{1F4BC}", // Work
  "\u{1F310}", // Internet
];

export default function CategoriesPage() {
  const { t } = useTranslation(["transaction", "common"]);
  const { workspaceId } = useWorkspaceContext();
  const { data: categories, isLoading } = useCategories(workspaceId);
  const createMutation = useCreateCategory(workspaceId);
  const updateMutation = useUpdateCategory(workspaceId);
  const deleteMutation = useDeleteCategory(workspaceId);

  const [dialogOpen, setDialogOpen] = useState(false);
  const [editing, setEditing] = useState<Category | null>(null);
  const [deleteId, setDeleteId] = useState<number | null>(null);
  const [name, setName] = useState("");
  const [icon, setIcon] = useState("");

  const openCreate = () => {
    setEditing(null);
    setName("");
    setIcon("");
    setDialogOpen(true);
  };

  const openEdit = (cat: Category) => {
    setEditing(cat);
    setName(cat.name);
    setIcon(cat.icon);
    setDialogOpen(true);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      if (editing) {
        await updateMutation.mutateAsync({ id: editing.id, name, icon });
        toast.success(t("transaction:categories.updated"));
      } else {
        await createMutation.mutateAsync({ name, icon });
        toast.success(t("transaction:categories.created"));
      }
      setDialogOpen(false);
    } catch (err) {
      toast.error(getErrorMessage(err, t("transaction:categories.saveFailed")));
    }
  };

  const handleDelete = async () => {
    if (deleteId === null) return;
    try {
      await deleteMutation.mutateAsync(deleteId);
      toast.success(t("transaction:categories.deleted"));
      setDeleteId(null);
    } catch (err) {
      toast.error(getErrorMessage(err, t("transaction:categories.deleteFailed")));
    }
  };

  const visibleCategories = categories?.filter((cat) => !isUncategorized(cat));

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">{t("transaction:categories.title")}</h1>
          <p className="text-sm text-muted-foreground">{t("transaction:categories.subtitle")}</p>
        </div>
        <Button size="sm" onClick={openCreate}>
          <Plus className="mr-1 h-4 w-4" /> {t("common:new")}
        </Button>
      </div>

      {isLoading ? (
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {[...Array(6)].map((_, i) => (
            <Skeleton key={i} className="h-20" />
          ))}
        </div>
      ) : !visibleCategories || visibleCategories.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <div className="rounded-full bg-muted p-3 mb-3">
            <Tags className="h-5 w-5 text-muted-foreground" />
          </div>
          <p className="text-sm font-medium">{t("transaction:categories.noCategoriesYet")}</p>
          <p className="text-sm text-muted-foreground mt-1">{t("transaction:createToGetStarted")}</p>
        </div>
      ) : (
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {visibleCategories.map((cat) => (
            <Card key={cat.id}>
              <CardContent className="flex items-center justify-between p-4">
                <div className="flex items-center gap-3">
                  <span className="text-xl">{cat.icon}</span>
                  <span className="font-medium">{getCategoryDisplayName(cat, t)}</span>
                </div>
                <div className="flex gap-1">
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => openEdit(cat)}
                  >
                    <Pencil className="h-4 w-4" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => setDeleteId(cat.id)}
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {editing ? t("transaction:categories.editTitle") : t("transaction:categories.newTitle")}
            </DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-1">
              <Label>{t("transaction:categories.name")}</Label>
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label>{t("transaction:categories.icon")}</Label>
              {icon && (
                <div className="flex items-center gap-2 text-sm text-muted-foreground">
                  <span className="text-2xl">{icon}</span>
                  <span>{t("transaction:categories.selected")}</span>
                </div>
              )}
              <div className="grid grid-cols-10 gap-1">
                {CATEGORY_ICONS.map((emoji) => (
                  <button
                    key={emoji}
                    type="button"
                    onClick={() => setIcon(emoji)}
                    className={`h-9 w-9 flex items-center justify-center rounded-md text-lg hover:bg-accent transition-colors ${
                      icon === emoji
                        ? "ring-2 ring-primary bg-accent"
                        : ""
                    }`}
                  >
                    {emoji}
                  </button>
                ))}
              </div>
              <Input
                value={icon}
                onChange={(e) => setIcon(e.target.value)}
                placeholder={t("transaction:categories.customEmoji")}
                className="mt-1"
              />
            </div>
            <DialogFooter>
              <Button
                type="submit"
                disabled={
                  createMutation.isPending || updateMutation.isPending
                }
              >
                {editing ? t("common:update") : t("common:create")}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => !open && setDeleteId(null)}
        title={t("transaction:categories.deleteTitle")}
        description={t("transaction:categories.deleteDescription")}
        onConfirm={handleDelete}
        loading={deleteMutation.isPending}
      />
    </div>
  );
}
