import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Trash2, X } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { MoneyDisplay } from "@/components/shared/MoneyDisplay";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog";
import {
  useBulkCategorizeTransactions,
  useBulkDeleteTransactions,
} from "@/api/transactions";
import { useCategories } from "@/api/categories";
import { useWorkspaceContext } from "@/hooks/useWorkspace";
import { getErrorMessage } from "@/lib/errors";
import type { Money } from "@/types";

interface Props {
  selectedIds: Set<number>;
  selectedTotals: Map<string, Money>;
  onClearSelection: () => void;
}

export function BulkActionBar({ selectedIds, selectedTotals, onClearSelection }: Props) {
  const { t } = useTranslation(["transaction", "common"]);
  const { workspaceId } = useWorkspaceContext();
  const { data: categories } = useCategories(workspaceId);
  const bulkCategorize = useBulkCategorizeTransactions(workspaceId);
  const bulkDelete = useBulkDeleteTransactions(workspaceId);

  const [categorizeOpen, setCategorizeOpen] = useState(false);
  const [selectedCategoryId, setSelectedCategoryId] = useState<string>("");
  const [deleteOpen, setDeleteOpen] = useState(false);

  const visible = selectedIds.size > 0;

  const handleCategorize = async () => {
    if (!selectedCategoryId) return;
    try {
      const result = await bulkCategorize.mutateAsync({
        ids: Array.from(selectedIds),
        category_id: parseInt(selectedCategoryId),
      });
      toast.success(t("transaction:bulk.updated", { count: result.updated }));
      setCategorizeOpen(false);
      setSelectedCategoryId("");
      onClearSelection();
    } catch (err) {
      toast.error(getErrorMessage(err, t("transaction:bulk.categorizeFailed")));
    }
  };

  const handleDelete = async () => {
    try {
      const result = await bulkDelete.mutateAsync(Array.from(selectedIds));
      toast.success(t("transaction:bulk.deleted", { count: result.deleted }));
      setDeleteOpen(false);
      onClearSelection();
    } catch (err) {
      toast.error(getErrorMessage(err, t("transaction:bulk.deleteFailed")));
    }
  };

  return (
    <>
      <div
        className={`fixed bottom-0 left-0 right-0 z-50 flex justify-center p-4 transition-transform duration-300 ease-out ${
          visible ? "translate-y-0" : "translate-y-full"
        }`}
      >
        <div className="flex items-center gap-4 rounded-lg border bg-background px-4 py-3 shadow-lg max-w-screen-xl w-full">
          <div className="flex items-center gap-3 flex-1 min-w-0">
            <span className="text-sm font-medium whitespace-nowrap">
              {selectedIds.size} {t("common:selected")}
            </span>
            <div className="h-4 w-px bg-border" />
            <div className="flex items-center gap-2 overflow-x-auto">
              {Array.from(selectedTotals.entries()).map(([currency, money]) => (
                <MoneyDisplay key={currency} value={money} />
              ))}
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Button size="sm" onClick={() => setCategorizeOpen(true)}>
              {t("transaction:bulk.categorize")}
            </Button>
            <Button
              size="sm"
              variant="destructive"
              onClick={() => setDeleteOpen(true)}
            >
              <Trash2 className="mr-1 h-4 w-4" />
              {t("common:delete")}
            </Button>
            <Button size="sm" variant="ghost" onClick={onClearSelection}>
              <X className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </div>

      <Dialog open={categorizeOpen} onOpenChange={setCategorizeOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {t("transaction:bulk.categorize")} {selectedIds.size} {t("transaction:bulk.transactions")}
            </DialogTitle>
          </DialogHeader>
          <Select value={selectedCategoryId} onValueChange={setSelectedCategoryId}>
            <SelectTrigger>
              <SelectValue placeholder={t("transaction:bulk.selectCategory")} />
            </SelectTrigger>
            <SelectContent>
              {categories?.map((c) => (
                <SelectItem key={c.id} value={String(c.id)}>
                  {c.icon} {c.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCategorizeOpen(false)}>
              {t("common:cancel")}
            </Button>
            <Button
              onClick={handleCategorize}
              disabled={!selectedCategoryId || bulkCategorize.isPending}
            >
              {bulkCategorize.isPending ? t("transaction:bulk.updating") : t("common:apply")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <ConfirmDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        title={`${t("common:delete")} ${selectedIds.size} ${t("transaction:bulk.transactions")}?`}
        description={t("transaction:bulk.deleteConfirm")}
        confirmLabel={t("transaction:bulk.deleteAll")}
        onConfirm={handleDelete}
        loading={bulkDelete.isPending}
      />
    </>
  );
}
