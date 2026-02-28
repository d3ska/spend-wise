import { useState, useEffect, useCallback, useMemo } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { Save } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useBudgets, useUpsertBudget, useDeleteBudget } from "@/api/fundings";
import { useCategories } from "@/api/categories";
import { useWorkspaceContext } from "@/hooks/useWorkspace";
import { getErrorMessage } from "@/lib/errors";
import { getCategoryDisplayName } from "@/lib/categoryUtils";
import type { Funding } from "@/types";

function currentMonth(): string {
  const d = new Date();
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}`;
}

export default function BudgetsPage() {
  const { t } = useTranslation(["transaction", "common"]);
  const { workspaceId } = useWorkspaceContext();
  const [month, setMonth] = useState(currentMonth);
  const { data: budgets, isLoading: budgetsLoading } = useBudgets(
    workspaceId,
    month,
  );
  const { data: categories, isLoading: catsLoading } =
    useCategories(workspaceId);
  const upsertMutation = useUpsertBudget(workspaceId);
  const deleteMutation = useDeleteBudget(workspaceId);

  const [overallAmount, setOverallAmount] = useState("");
  const [categoryAmounts, setCategoryAmounts] = useState<
    Record<number, string>
  >({});
  const [saving, setSaving] = useState(false);

  // Sync form state when budgets load or month changes.
  const syncForm = useCallback(
    (data: Funding[] | undefined) => {
      if (!data) {
        setOverallAmount("");
        setCategoryAmounts({});
        return;
      }
      const overall = data.find((f) => f.category_id === null);
      setOverallAmount(overall ? parseFloat(overall.amount.amount).toString() : "");
      const catMap: Record<number, string> = {};
      for (const f of data) {
        if (f.category_id !== null) {
          catMap[f.category_id] = parseFloat(f.amount.amount).toString();
        }
      }
      setCategoryAmounts(catMap);
    },
    [],
  );

  useEffect(() => {
    syncForm(budgets);
  }, [budgets, syncForm]);

  const handleSave = async () => {
    setSaving(true);
    try {
      const deleteIds: number[] = [];
      const upserts: {
        year_month: string;
        category_id: number | null;
        amount: { amount: string; currency: string };
      }[] = [];

      // Overall budget
      const existingOverall = budgets?.find((f) => f.category_id === null);
      if (overallAmount.trim() !== "") {
        const amt = parseFloat(overallAmount);
        if (!isNaN(amt) && amt > 0) {
          upserts.push({
            year_month: month,
            category_id: null,
            amount: { amount: amt.toFixed(2), currency: "PLN" },
          });
        }
      } else if (existingOverall) {
        deleteIds.push(existingOverall.id);
      }

      // Per-category budgets
      for (const cat of categories ?? []) {
        const val = categoryAmounts[cat.id] ?? "";
        const existing = budgets?.find((f) => f.category_id === cat.id);
        if (val.trim() !== "") {
          const amt = parseFloat(val);
          if (!isNaN(amt) && amt > 0) {
            upserts.push({
              year_month: month,
              category_id: cat.id,
              amount: { amount: amt.toFixed(2), currency: "PLN" },
            });
          }
        } else if (existing) {
          deleteIds.push(existing.id);
        }
      }

      // Run deletes first so validation sees correct DB state for upserts.
      for (const id of deleteIds) {
        await deleteMutation.mutateAsync(id);
      }
      for (const body of upserts) {
        await upsertMutation.mutateAsync(body);
      }

      toast.success(t("transaction:budgets.saved"));
    } catch (err) {
      toast.error(getErrorMessage(err, t("transaction:budgets.saveFailed")));
    } finally {
      setSaving(false);
    }
  };

  const isLoading = budgetsLoading || catsLoading;

  const categorySum = useMemo(() => {
    let sum = 0;
    for (const val of Object.values(categoryAmounts)) {
      const n = parseFloat(val);
      if (!isNaN(n) && n > 0) sum += n;
    }
    return sum;
  }, [categoryAmounts]);

  const overallNum = parseFloat(overallAmount) || 0;
  const budgetExceeded = overallNum > 0 && categorySum > overallNum;
  const remaining = overallNum > 0 ? overallNum - categorySum : null;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">{t("transaction:budgets.title")}</h1>
          <p className="text-sm text-muted-foreground">{t("transaction:budgets.subtitle")}</p>
        </div>
        <div className="flex items-center gap-2">
          <Input
            type="month"
            value={month}
            onChange={(e) => setMonth(e.target.value)}
            className="w-44"
          />
          <Button size="sm" onClick={handleSave} disabled={saving || isLoading || budgetExceeded}>
            <Save className="mr-1 h-4 w-4" />
            {saving ? t("common:saving") : t("common:save")}
          </Button>
        </div>
      </div>

      {isLoading ? (
        <div className="space-y-2">
          {[...Array(4)].map((_, i) => (
            <Skeleton key={i} className="h-12 w-full" />
          ))}
        </div>
      ) : (
        <div className="space-y-4">
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-base">{t("transaction:budgets.overallBudget")}</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex items-center gap-2">
                <Label className="sr-only">{t("transaction:budgets.overallBudgetAmount")}</Label>
                <Input
                  type="text"
                  inputMode="decimal"
                  placeholder={t("transaction:budgets.noOverallBudget")}
                  value={overallAmount}
                  onChange={(e) => setOverallAmount(e.target.value)}
                  className="max-w-xs"
                />
                <span className="text-muted-foreground text-sm">PLN</span>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-base">{t("transaction:budgets.categoryBudgets")}</CardTitle>
              {remaining !== null && (
                <p
                  className={`text-sm ${budgetExceeded ? "text-red-500" : "text-muted-foreground"}`}
                >
                  {budgetExceeded
                    ? `${t("transaction:budgets.exceeds")} ${(categorySum - overallNum).toFixed(2)} PLN`
                    : `${remaining.toFixed(2)} ${t("transaction:budgets.remaining")}`}
                </p>
              )}
            </CardHeader>
            <CardContent>
              {!categories || categories.length === 0 ? (
                <p className="text-muted-foreground text-sm">
                  {t("transaction:budgets.noCategoriesYet")}
                </p>
              ) : (
                <div className="space-y-3">
                  {categories.map((cat) => (
                    <div
                      key={cat.id}
                      className="flex items-center gap-3"
                    >
                      <span className="w-6 text-center">{cat.icon}</span>
                      <span className="w-32 truncate text-sm font-medium">
                        {getCategoryDisplayName(cat, t)}
                      </span>
                      <Input
                        type="text"
                        inputMode="decimal"
                        placeholder={t("transaction:budgets.noBudget")}
                        value={categoryAmounts[cat.id] ?? ""}
                        onChange={(e) =>
                          setCategoryAmounts((prev) => ({
                            ...prev,
                            [cat.id]: e.target.value,
                          }))
                        }
                        className="max-w-xs"
                      />
                      <span className="text-muted-foreground text-sm">
                        PLN
                      </span>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
}
