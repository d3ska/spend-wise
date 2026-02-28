import { useTranslation } from "react-i18next";
import type { Money } from "@/types";
import { formatMoney } from "@/lib/money";
import { cn } from "@/lib/utils";

interface BudgetProgressBarProps {
  spent: Money;
  budget: Money;
  size?: "sm" | "lg";
  className?: string;
}

export function BudgetProgressBar({
  spent,
  budget,
  size = "sm",
  className,
}: BudgetProgressBarProps) {
  const { t } = useTranslation(["common"]);
  const spentNum = parseFloat(spent.amount);
  const budgetNum = parseFloat(budget.amount);
  const ratio = budgetNum > 0 ? spentNum / budgetNum : 0;
  const pct = Math.min(ratio * 100, 100);
  const overBudget = ratio > 1;
  const remainingNum = Math.abs(budgetNum - spentNum);
  const remainingMoney = formatMoney({
    amount: String(remainingNum),
    currency: spent.currency,
  });

  const barColor =
    ratio > 1
      ? "bg-red-500"
      : ratio >= 0.75
        ? "bg-yellow-500"
        : "bg-emerald-500";

  const pctColor =
    ratio > 1
      ? "text-red-500"
      : ratio >= 0.75
        ? "text-yellow-600 dark:text-yellow-500"
        : "text-emerald-600 dark:text-emerald-500";

  const isLg = size === "lg";

  return (
    <div className={cn("space-y-2", className)}>
      <div className="flex items-baseline justify-between gap-2">
        <div className="min-w-0">
          <span
            className={cn(
              "font-bold tabular-nums",
              isLg ? "text-3xl" : "text-lg",
            )}
          >
            {formatMoney(spent)}
          </span>
          <span
            className={cn(
              "text-muted-foreground ml-1.5",
              isLg ? "text-sm" : "text-xs",
            )}
          >
            {t("common:of")} {formatMoney(budget)}
          </span>
        </div>
        <span
          className={cn(
            "font-semibold tabular-nums shrink-0",
            isLg ? "text-base" : "text-sm",
            pctColor,
          )}
        >
          {Math.round(ratio * 100)}%
        </span>
      </div>
      <div className={cn("w-full rounded-full bg-muted", isLg ? "h-3" : "h-2.5")}>
        <div
          className={cn("h-full rounded-full transition-all", barColor)}
          style={{ width: `${pct}%` }}
        />
      </div>
      <p
        className={cn(
          "tabular-nums",
          isLg ? "text-xs" : "text-[11px]",
          overBudget ? "text-red-500" : "text-muted-foreground",
        )}
      >
        {overBudget
          ? `${remainingMoney} ${t("common:overBudget")}`
          : `${remainingMoney} ${t("common:remaining")}`}
      </p>
    </div>
  );
}
