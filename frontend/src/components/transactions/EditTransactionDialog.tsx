import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { useCategories } from "@/api/categories";
import { useUpdateTransaction } from "@/api/transactions";
import { useWorkspaceContext } from "@/hooks/useWorkspace";
import { getErrorMessage } from "@/lib/errors";
import { getCategoryDisplayName } from "@/lib/categoryUtils";
import type { Transaction, TransactionType } from "@/types";

const txSchema = z.object({
  description: z.string().min(1),
  date: z.string().min(1),
  type: z.enum(["expense", "income", "transfer"]),
  notes: z.string().optional(),
  category_id: z.string(),
});

type TxFormValues = z.infer<typeof txSchema>;

interface Props {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  transaction: Transaction;
}

export function EditTransactionDialog({
  open,
  onOpenChange,
  transaction,
}: Props) {
  const { t } = useTranslation(["transaction", "common"]);
  const { workspaceId } = useWorkspaceContext();
  const { data: categories } = useCategories(workspaceId);
  const updateMutation = useUpdateTransaction(workspaceId);
  const [initialized, setInitialized] = useState(false);

  const form = useForm<TxFormValues>({
    resolver: zodResolver(txSchema),
    defaultValues: {
      description: "",
      date: "",
      type: "expense",
      notes: "",
      category_id: "",
    },
  });

  useEffect(() => {
    if (open && !initialized) {
      const entry = transaction.entries[0];
      form.reset({
        description: transaction.description,
        date: transaction.date,
        type: transaction.type,
        notes: transaction.notes ?? "",
        category_id: entry?.category_id ? String(entry.category_id) : "",
      });
      setInitialized(true);
    }
    if (!open) {
      setInitialized(false);
    }
  }, [open, initialized, transaction, form]);

  const onSubmit = async (values: TxFormValues) => {
    const parsedCategoryId = values.category_id
      ? parseInt(values.category_id)
      : undefined;
    const currency = transaction.total_amount.currency;
    const totalAmount = transaction.total_amount.amount;

    try {
      await updateMutation.mutateAsync({
        txId: transaction.id,
        body: {
          description: values.description,
          date: values.date,
          type: values.type as TransactionType,
          notes: values.notes ?? "",
          total_amount: { amount: totalAmount, currency },
          entries: [
            {
              category_id: parsedCategoryId,
              participant_id: null,
              amount: { amount: totalAmount, currency },
            },
          ],
        },
      });
      toast.success(t("transaction:updated"));
      onOpenChange(false);
    } catch (err) {
      toast.error(getErrorMessage(err, t("transaction:updateFailed")));
    }
  };

  const txType = form.watch("type");
  const categoryId = form.watch("category_id");

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{t("transaction:editTitle")}</DialogTitle>
        </DialogHeader>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
          <div className="space-y-1">
            <Label>{t("transaction:description")}</Label>
            <Input {...form.register("description")} />
          </div>

          <div className="space-y-1">
            <Label>{t("transaction:type")}</Label>
            <Select
              value={form.watch("type")}
              onValueChange={(v) => {
                const newType = v as "expense" | "income" | "transfer";
                form.setValue("type", newType);
                if (newType !== "expense") {
                  form.setValue("category_id", "");
                }
              }}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="expense">{t("transaction:expense")}</SelectItem>
                <SelectItem value="income">{t("transaction:income")}</SelectItem>
                <SelectItem value="transfer">{t("transaction:transfer")}</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-1">
            <Label>{t("transaction:date")}</Label>
            <Input type="date" {...form.register("date")} />
          </div>

          {txType === "expense" && (
            <div className="space-y-1">
              <Label>{t("transaction:category")}</Label>
              <Select
                value={categoryId || undefined}
                onValueChange={(v) => form.setValue("category_id", v)}
              >
                <SelectTrigger>
                  <SelectValue placeholder={t("transaction:selectCategoryOptional")} />
                </SelectTrigger>
                <SelectContent>
                  {categories?.map((c) => (
                    <SelectItem key={c.id} value={String(c.id)}>
                      {c.icon} {getCategoryDisplayName(c, t)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          )}

          <div className="space-y-1">
            <Label>{t("transaction:notes")}</Label>
            <Input {...form.register("notes")} />
          </div>

          <DialogFooter>
            <Button type="submit" disabled={updateMutation.isPending}>
              {updateMutation.isPending ? t("common:saving") : t("common:save")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
