import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { AlertTriangle, ArrowDown, ArrowUp, Receipt, Wallet } from "lucide-react";
import { useDateRange } from "@/hooks/useDateRange";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Skeleton } from "@/components/ui/skeleton";
import { AreaChart, Area, XAxis, YAxis, PieChart, Pie, Cell, ResponsiveContainer, Tooltip } from "recharts";
import { useSummary } from "@/api/summary";
import { useTransactions } from "@/api/transactions";
import { useListConnections } from "@/api/bank-connections";
import { useWorkspaceContext } from "@/hooks/useWorkspace";
import { DateRangePicker } from "@/components/shared/DateRangePicker";
import { MoneyDisplay } from "@/components/shared/MoneyDisplay";
import { BudgetProgressBar } from "@/components/shared/BudgetProgressBar";
import { formatDate, formatDateShort } from "@/lib/dates";
import { formatMoney } from "@/lib/money";
import { getCategoryDisplayName, isUncategorized } from "@/lib/categoryUtils";

const COLORS = [
  "var(--chart-1)",
  "var(--chart-2)",
  "var(--chart-3)",
  "var(--chart-4)",
  "var(--chart-5)",
];

function PercentChange({ current, previous, newLabel }: { current: number; previous: number; newLabel: string }) {
  if (previous === 0 && current === 0) return null;
  if (previous === 0) {
    return (
      <span className="flex items-center gap-0.5 text-xs font-medium text-emerald-600">
        <ArrowUp className="h-3 w-3" />{newLabel}
      </span>
    );
  }
  const pct = ((current - previous) / previous) * 100;
  if (pct === 0) return null;
  const isUp = pct > 0;
  return (
    <span className={`flex items-center gap-0.5 text-xs font-medium ${isUp ? "text-rose-600" : "text-emerald-600"}`}>
      {isUp ? <ArrowUp className="h-3 w-3" /> : <ArrowDown className="h-3 w-3" />}
      {Math.abs(pct).toFixed(0)}%
    </span>
  );
}

export default function DashboardPage() {
  const { t } = useTranslation(["transaction", "common", "workspace"]);
  const navigate = useNavigate();
  const { workspaceId } = useWorkspaceContext();
  const { range, setRange } = useDateRange();
  const [expiredDismissed, setExpiredDismissed] = useState(false);

  const { data: connections } = useListConnections();
  const expiredConnections = connections?.filter(
    (c) => c.expiry_status === "expired",
  ) ?? [];

  const { data: summary, isLoading: summaryLoading } = useSummary(
    workspaceId,
    range.from,
    range.to,
  );

  const { data: recentTxs, isLoading: txLoading } = useTransactions(
    workspaceId,
    { from: range.from, to: range.to, limit: 5, offset: 0, type: "expense" },
  );

  const chartData =
    summary?.by_category
      .filter((c) => !isUncategorized(c))
      .map((c) => ({
        name: getCategoryDisplayName(c, t),
        value: parseFloat(c.spent.amount),
        category_id: c.category_id,
      })) ?? [];

  const trendData =
    summary?.daily_trend.map((d) => ({
      date: formatDateShort(new Date(d.date)),
      fullDate: formatDate(d.date),
      value: parseFloat(d.spent.amount),
    })) ?? [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">{t("transaction:dashboard.title")}</h1>
          <p className="text-sm text-muted-foreground">{t("transaction:dashboard.subtitle")}</p>
        </div>
        <DateRangePicker value={range} onChange={setRange} />
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              {t("transaction:dashboard.totalSpent")}
            </CardTitle>
            <div className="bg-primary/10 p-2 rounded-md">
              <Wallet className="h-4 w-4 text-primary" />
            </div>
          </CardHeader>
          <CardContent>
            {summaryLoading ? (
              <Skeleton className="h-8 w-32" />
            ) : summary ? (
              <>
                {summary.overall_budget ? (
                  <BudgetProgressBar
                    spent={summary.total_spent}
                    budget={summary.overall_budget}
                    size="lg"
                  />
                ) : (
                  <MoneyDisplay
                    value={summary.total_spent}
                    className="text-3xl font-bold"
                  />
                )}
                <PercentChange
                  current={parseFloat(summary.total_spent.amount)}
                  previous={parseFloat(summary.prev_total_spent.amount)}
                  newLabel={t("transaction:new")}
                />
              </>
            ) : (
              <p className="text-muted-foreground">{t("common:noDataYet")}</p>
            )}
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              {t("transaction:dashboard.transactionCount")}
            </CardTitle>
            <div className="bg-primary/10 p-2 rounded-md">
              <Receipt className="h-4 w-4 text-primary" />
            </div>
          </CardHeader>
          <CardContent>
            {summaryLoading ? (
              <Skeleton className="h-8 w-16" />
            ) : (
              <div className="flex items-baseline gap-2">
                <p className="text-3xl font-bold">
                  {summary?.transaction_count ?? 0}
                </p>
                {summary && (
                  <PercentChange
                    current={summary.transaction_count}
                    previous={summary.prev_transaction_count}
                    newLabel={t("transaction:new")}
                  />
                )}
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">{t("transaction:dashboard.spendingTrend")}</CardTitle>
        </CardHeader>
        <CardContent>
          {summaryLoading ? (
            <Skeleton className="h-48 w-full" />
          ) : trendData.length === 0 ? (
            <p className="text-muted-foreground text-sm py-8 text-center">
              {t("transaction:dashboard.noSpendingData")}
            </p>
          ) : (
            <ResponsiveContainer width="100%" height={200}>
              <AreaChart data={trendData}>
                <XAxis
                  dataKey="date"
                  tick={{ fontSize: 12 }}
                  tickLine={false}
                  axisLine={false}
                />
                <YAxis
                  tick={{ fontSize: 12 }}
                  tickLine={false}
                  axisLine={false}
                  width={60}
                />
                <Tooltip
                  cursor={{ stroke: "var(--chart-1)", strokeOpacity: 0.3 }}
                  content={({ active, payload }) => {
                    if (!active || !payload?.length) return null;
                    const d = payload[0].payload as { date: string; fullDate: string; value: number };
                    return (
                      <div className="rounded-md border bg-popover px-3 py-1.5 shadow-md">
                        <p className="text-xs text-muted-foreground">{d.fullDate}</p>
                        <p className="text-sm font-semibold">
                          {formatMoney({
                            amount: String(d.value),
                            currency: summary?.total_spent.currency ?? "PLN",
                          })}
                        </p>
                      </div>
                    );
                  }}
                />
                <Area
                  type="monotone"
                  dataKey="value"
                  stroke="var(--chart-1)"
                  fill="var(--chart-1)"
                  fillOpacity={0.2}
                />
              </AreaChart>
            </ResponsiveContainer>
          )}
        </CardContent>
      </Card>

      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">{t("transaction:dashboard.spendingByCategory")}</CardTitle>
          </CardHeader>
          <CardContent>
            {summaryLoading ? (
              <Skeleton className="h-48 w-full" />
            ) : chartData.length === 0 ? (
              <p className="text-muted-foreground text-sm py-8 text-center">
                {t("transaction:dashboard.noSpendingData")}
              </p>
            ) : (
              <div className="space-y-4">
                <ResponsiveContainer width="100%" height={200}>
                  <PieChart>
                    <Pie
                      data={chartData}
                      dataKey="value"
                      nameKey="name"
                      cx="50%"
                      cy="50%"
                      innerRadius={50}
                      outerRadius={80}
                    >
                      {chartData.map((entry, i) => (
                        <Cell
                          key={i}
                          fill={COLORS[i % COLORS.length]}
                          className="cursor-pointer"
                          onClick={() => navigate(`/transactions?category_id=${entry.category_id}`)}
                        />
                      ))}
                    </Pie>
                    <Tooltip
                      content={({ active, payload }) => {
                        if (!active || !payload?.length) return null;
                        const d = payload[0] as { name: string; value: number };
                        return (
                          <div className="rounded-md border bg-popover px-3 py-1.5 shadow-md">
                            <p className="text-xs text-muted-foreground">{d.name}</p>
                            <p className="text-sm font-semibold">
                              {formatMoney({
                                amount: String(d.value),
                                currency: summary?.total_spent.currency ?? "PLN",
                              })}
                            </p>
                          </div>
                        );
                      }}
                    />
                  </PieChart>
                </ResponsiveContainer>
                <div className="space-y-2">
                  {summary!.by_category.filter((c) => !isUncategorized(c)).map((c, i) => (
                    <div
                      key={c.category_id}
                      className="flex items-center justify-between text-sm cursor-pointer hover:bg-muted/50 rounded-md px-1 -mx-1 py-0.5"
                      onClick={() => navigate(`/transactions?category_id=${c.category_id}`)}
                    >
                      <div className="flex items-center gap-2">
                        <div
                          className="h-2.5 w-2.5 rounded-full shrink-0"
                          style={{ backgroundColor: COLORS[i % COLORS.length] }}
                        />
                        <span>{c.icon}</span>
                        <span className="truncate">{getCategoryDisplayName(c, t)}</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <span className="font-medium">
                          {formatMoney(c.spent)}
                        </span>
                        <PercentChange
                          current={parseFloat(c.spent.amount)}
                          previous={parseFloat(c.prev_spent.amount)}
                          newLabel={t("transaction:new")}
                        />
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle className="text-base">{t("transaction:dashboard.recentTransactions")}</CardTitle>
            <Link
              to="/transactions"
              className="text-sm text-muted-foreground hover:underline"
            >
              {t("common:viewAll")}
            </Link>
          </CardHeader>
          <CardContent>
            {txLoading ? (
              <div className="space-y-3">
                {[...Array(3)].map((_, i) => (
                  <Skeleton key={i} className="h-6 w-full" />
                ))}
              </div>
            ) : !recentTxs || recentTxs.length === 0 ? (
              <p className="text-muted-foreground text-sm py-4 text-center">
                {t("transaction:dashboard.noTransactionsYet")}
              </p>
            ) : (
              <div className="space-y-3">
                {recentTxs.map((tx) => (
                  <div
                    key={tx.id}
                    className="flex items-center justify-between text-sm"
                  >
                    <div>
                      <p className="font-medium">{tx.description}</p>
                      <p className="text-muted-foreground text-xs">
                        {formatDate(tx.date)}
                      </p>
                    </div>
                    <MoneyDisplay value={tx.total_amount} />
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {summary &&
        summary.by_category.some((c) => c.budget !== null) && (
          <Card>
            <CardHeader>
              <CardTitle className="text-base">{t("transaction:dashboard.categoryBudgets")}</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {summary.by_category
                .filter((c) => c.budget !== null)
                .map((c) => (
                  <div key={c.category_id} className="space-y-1">
                    <div className="flex items-center gap-2 text-sm font-medium">
                      <span>{c.icon}</span>
                      <span>{getCategoryDisplayName(c, t)}</span>
                    </div>
                    <BudgetProgressBar spent={c.spent} budget={c.budget!} />
                  </div>
                ))}
            </CardContent>
          </Card>
        )}

      <Dialog
        open={expiredConnections.length > 0 && !expiredDismissed}
        onOpenChange={(open) => {
          if (!open) setExpiredDismissed(true);
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <AlertTriangle className="h-5 w-5 text-destructive" />
              {t("workspace:bank.expiredWarning")}
            </DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">
            {expiredConnections.length === 1
              ? t("workspace:bank.singleExpired", { name: expiredConnections[0].institution_name })
              : t("workspace:bank.multipleExpired", { count: expiredConnections.length })}{" "}
            {t("workspace:bank.expiredDescription")}
          </p>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setExpiredDismissed(true)}
            >
              {t("common:dismiss")}
            </Button>
            <Button
              onClick={() => navigate("/settings")}
            >
              {t("workspace:bank.goToSettings")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
