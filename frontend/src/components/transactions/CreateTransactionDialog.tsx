import { useEffect, useMemo, useRef, useState } from "react";
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
import { useCreateTransaction } from "@/api/transactions";
import { useRules } from "@/api/rules";
import { useWorkspaceContext } from "@/hooks/useWorkspace";
import { getErrorMessage } from "@/lib/errors";
import { getCategoryDisplayName, isUncategorized } from "@/lib/categoryUtils";
import type { Category, Rule, TransactionType } from "@/types";

const txSchema = z.object({
  description: z.string().min(1),
  date: z.string().min(1),
  type: z.enum(["expense", "income", "transfer"]),
  notes: z.string().optional(),
  amount: z.string().min(1),
  currency: z.string().min(1),
  category_id: z.string(),
});

type TxFormValues = z.infer<typeof txSchema>;

interface Props {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

function matchCategory(
  description: string,
  rules: Rule[],
): number | null {
  if (!description.trim()) return null;
  const lower = description.toLowerCase();
  const enabledRules = rules
    .filter((r) => r.enabled)
    .sort((a, b) => b.priority - a.priority);
  for (const rule of enabledRules) {
    if (lower.includes(rule.match_pattern.toLowerCase())) {
      return rule.target_category_id;
    }
  }
  return null;
}

export function CreateTransactionDialog({ open, onOpenChange }: Props) {
  const { t } = useTranslation(["transaction", "common"]);
  const { workspaceId } = useWorkspaceContext();
  const { data: categories } = useCategories(workspaceId);
  const { data: rules } = useRules(workspaceId);
  const createMutation = useCreateTransaction(workspaceId);

  const [categoryOverridden, setCategoryOverridden] = useState(false);
  const prevSuggestionRef = useRef<number | null>(null);

  const form = useForm<TxFormValues>({
    resolver: zodResolver(txSchema),
    defaultValues: {
      description: "",
      date: "",
      type: "expense",
      notes: "",
      amount: "",
      currency: "PLN",
      category_id: "",
    },
  });

  const txType = form.watch("type");
  const description = form.watch("description");
  const categoryId = form.watch("category_id");

  // Auto-suggest category from rules (only for expenses)
  useEffect(() => {
    if (categoryOverridden || !rules || txType !== "expense") return;
    const matched = matchCategory(description, rules);
    if (matched !== null && matched !== prevSuggestionRef.current) {
      prevSuggestionRef.current = matched;
      form.setValue("category_id", String(matched));
    } else if (matched === null && prevSuggestionRef.current !== null) {
      prevSuggestionRef.current = null;
      form.setValue("category_id", "");
    }
  }, [description, rules, categoryOverridden, txType, form]);

  const handleCategoryChange = (value: string) => {
    setCategoryOverridden(true);
    form.setValue("category_id", value);
  };

  const onSubmit = async (values: TxFormValues) => {
    const parsedCategoryId = values.category_id
      ? parseInt(values.category_id)
      : undefined;
    const currency = values.currency;
    const totalAmount = parseFloat(values.amount);

    try {
      await createMutation.mutateAsync({
        description: values.description,
        date: values.date,
        type: values.type as TransactionType,
        notes: values.notes ?? "",
        total_amount: { amount: totalAmount.toFixed(2), currency },
        entries: [
          {
            category_id: parsedCategoryId,
            participant_id: null,
            amount: { amount: totalAmount.toFixed(2), currency },
          },
        ],
      });
      toast.success(t("transaction:created"));
      onOpenChange(false);
      resetForm();
    } catch (err) {
      toast.error(getErrorMessage(err, t("transaction:createFailed")));
    }
  };

  const resetForm = () => {
    form.reset();
    setCategoryOverridden(false);
    prevSuggestionRef.current = null;
  };

  const handleOpenChange = (open: boolean) => {
    if (!open) resetForm();
    onOpenChange(open);
  };

  const suggestedCategory: Category | undefined = useMemo(() => {
    if (!rules || !categories) return undefined;
    const matched = matchCategory(description, rules);
    if (matched === null) return undefined;
    return categories.find((c) => c.id === matched);
  }, [description, rules, categories]);

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="max-w-lg max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{t("transaction:createTitle")}</DialogTitle>
        </DialogHeader>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
          {/* Description */}
          <div className="space-y-1">
            <Label>{t("transaction:description")}</Label>
            <Input {...form.register("description")} />
            {suggestedCategory && !categoryOverridden && txType === "expense" && (
              <p className="text-xs text-muted-foreground">
                {t("transaction:autoMatchedCategory")} {getCategoryDisplayName(suggestedCategory, t)}
              </p>
            )}
          </div>

          {/* Type */}
          <div className="space-y-1">
            <Label>{t("transaction:type")}</Label>
            <Select
              value={form.watch("type")}
              onValueChange={(v) => {
                const newType = v as "expense" | "income" | "transfer";
                form.setValue("type", newType);
                if (newType !== "expense") {
                  form.setValue("category_id", "");
                  setCategoryOverridden(false);
                  prevSuggestionRef.current = null;
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

          {/* Date + Amount */}
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-1">
              <Label>{t("transaction:date")}</Label>
              <Input type="date" {...form.register("date")} />
            </div>
            <div className="space-y-1">
              <Label>{t("transaction:amount")}</Label>
              <Input
                type="number"
                step="0.01"
                {...form.register("amount")}
              />
            </div>
          </div>

          {/* Category (only for expenses) */}
          {txType === "expense" && (
            <div className="space-y-1">
              <Label>{t("transaction:category")}</Label>
              <Select
                value={categoryId || undefined}
                onValueChange={handleCategoryChange}
              >
                <SelectTrigger>
                  <SelectValue placeholder={t("transaction:selectCategory")} />
                </SelectTrigger>
                <SelectContent>
                  {categories?.filter((c) => !isUncategorized(c)).map((c) => (
                    <SelectItem key={c.id} value={String(c.id)}>
                      {c.icon} {getCategoryDisplayName(c, t)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {(!categories || categories.length === 0) && (
                <p className="text-xs text-muted-foreground">
                  {t("transaction:noCategoriesYet")}
                </p>
              )}
            </div>
          )}

          {/* Notes */}
          <div className="space-y-1">
            <Label>{t("transaction:notes")}</Label>
            <Input {...form.register("notes")} />
          </div>

          <DialogFooter>
            <Button
              type="submit"
              disabled={createMutation.isPending}
            >
              {createMutation.isPending ? t("common:creating") : t("common:create")}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
