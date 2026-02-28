import { useCallback, useEffect, useMemo, useState, type KeyboardEvent } from "react";
import { useSearchParams } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { ArrowDown, ArrowLeftRight, ArrowUp, Building2, Pencil, Plus, Receipt, Trash2 } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Checkbox } from "@/components/ui/checkbox";
import { Skeleton } from "@/components/ui/skeleton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog";
import { DateRangePicker } from "@/components/shared/DateRangePicker";
import { MoneyDisplay } from "@/components/shared/MoneyDisplay";
import { CreateTransactionDialog } from "@/components/transactions/CreateTransactionDialog";
import { BulkActionBar } from "@/components/transactions/BulkActionBar";
import { EditTransactionDialog } from "@/components/transactions/EditTransactionDialog";
import { TransactionDetailDialog } from "@/components/transactions/TransactionDetailDialog";
import {
  useTransactions,
  useDeleteTransaction,
} from "@/api/transactions";
import { useCategories } from "@/api/categories";
import { useLinkedBankAccounts } from "@/api/bank-accounts";
import { useListMembers } from "@/api/members";
import { useWorkspaceContext } from "@/hooks/useWorkspace";
import { useAuth } from "@/hooks/useAuth";
import { useDateRange } from "@/hooks/useDateRange";
import { formatDate } from "@/lib/dates";
import { getErrorMessage } from "@/lib/errors";
import { getCategoryDisplayName } from "@/lib/categoryUtils";
import { MemberAvatar } from "@/components/shared/MemberAvatar";
import type { Money, Transaction, TransactionType } from "@/types";

const PAGE_SIZE = 50;

export default function TransactionsPage() {
  const { t } = useTranslation(["transaction", "common", "workspace"]);
  const { workspaceId, workspace } = useWorkspaceContext();
  const { user: currentUser } = useAuth();
  const [searchParams, setSearchParams] = useSearchParams();
  const { range, setRange } = useDateRange();
  const [page, setPage] = useState(0);
  const [createOpen, setCreateOpen] = useState(false);
  const [deleteId, setDeleteId] = useState<number | null>(null);
  const [editTx, setEditTx] = useState<Transaction | null>(null);
  const [detailTxId, setDetailTxId] = useState<number | null>(null);
  const [typeFilter, setTypeFilter] = useState<TransactionType | "all">("expense");
  const [currencyFilter, setCurrencyFilter] = useState<string>("all");
  const [categoryFilter, setCategoryFilter] = useState<string>(() => {
    const param = searchParams.get("category_id");
    if (param) {
      // Clear the search param so it doesn't stick on subsequent filter changes.
      setSearchParams({}, { replace: true });
      return param;
    }
    return "all";
  });
  const [bankFilter, setBankFilter] = useState<string>("all");
  const [amountMin, setAmountMin] = useState("");
  const [amountMax, setAmountMax] = useState("");
  const [appliedAmountMin, setAppliedAmountMin] = useState("");
  const [appliedAmountMax, setAppliedAmountMax] = useState("");
  const [sortBy, setSortBy] = useState<"date" | "amount" | "description">("date");
  const [sortOrder, setSortOrder] = useState<"asc" | "desc">("desc");
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set());

  const { data: categories } = useCategories(workspaceId);
  const { data: linkedBankAccounts } = useLinkedBankAccounts(workspaceId);
  const isMultiMember = (workspace?.member_count ?? 0) > 1;
  const { data: members } = useListMembers(workspaceId);

  const { data: txs, isLoading } = useTransactions(workspaceId, {
    from: range.from,
    to: range.to,
    limit: PAGE_SIZE,
    offset: page * PAGE_SIZE,
    ...(typeFilter !== "all" && { type: typeFilter }),
    ...(currencyFilter !== "all" && { currency: currencyFilter }),
    ...(categoryFilter !== "all" && { category_id: parseInt(categoryFilter) }),
    ...(bankFilter !== "all" && { bank_account_id: parseInt(bankFilter) }),
    ...(appliedAmountMin && { amount_min: appliedAmountMin }),
    ...(appliedAmountMax && { amount_max: appliedAmountMax }),
    sort_by: sortBy,
    sort_order: sortOrder,
  });

  const categoryMap = useMemo(() => {
    const map = new Map<number, { name: string; icon: string; slug: string | null }>();
    categories?.forEach((c) => map.set(c.id, { name: c.name, icon: c.icon, slug: c.slug }));
    return map;
  }, [categories]);

  const memberMap = useMemo(() => {
    const map = new Map<number, { display_name: string; avatar_url: string }>();
    members?.forEach((m) => map.set(m.user_id, { display_name: m.display_name, avatar_url: m.avatar_url }));
    return map;
  }, [members]);
  const deleteMutation = useDeleteTransaction(workspaceId);
  // Clear selection when filters, page, sort, or date range change.
  useEffect(() => {
    setSelectedIds(new Set());
  }, [page, typeFilter, currencyFilter, categoryFilter, bankFilter, appliedAmountMin, appliedAmountMax, sortBy, sortOrder, range.from, range.to]);

  const toggleSelection = useCallback((id: number) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }, []);

  const toggleSelectAll = useCallback(() => {
    if (!txs) return;
    setSelectedIds((prev) => {
      if (prev.size === txs.length) return new Set();
      return new Set(txs.map((tx) => tx.id));
    });
  }, [txs]);

  const clearSelection = useCallback(() => setSelectedIds(new Set()), []);

  const selectedTotals = useMemo(() => {
    const totals = new Map<string, Money>();
    if (!txs) return totals;
    for (const tx of txs) {
      if (!selectedIds.has(tx.id)) continue;
      const currency = tx.total_amount.currency;
      const existing = totals.get(currency);
      if (existing) {
        totals.set(currency, {
          amount: (parseFloat(existing.amount) + parseFloat(tx.total_amount.amount)).toFixed(2),
          currency,
        });
      } else {
        totals.set(currency, { ...tx.total_amount });
      }
    }
    return totals;
  }, [txs, selectedIds]);

  const allSelected = !!txs && txs.length > 0 && selectedIds.size === txs.length;
  const someSelected = selectedIds.size > 0 && !allSelected;

  const handleDelete = async () => {
    if (deleteId === null) return;
    try {
      await deleteMutation.mutateAsync(deleteId);
      toast.success(t("transaction:deleted"));
      setDeleteId(null);
    } catch (err) {
      toast.error(getErrorMessage(err, t("transaction:deleteFailed")));
    }
  };

  return (
    <div className={`space-y-6 ${selectedIds.size > 0 ? "pb-20" : ""}`}>
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">{t("transaction:title")}</h1>
          <p className="text-sm text-muted-foreground">{t("transaction:subtitle")}</p>
        </div>
        <div className="flex items-center gap-2">
          <Button size="sm" onClick={() => setCreateOpen(true)}>
            <Plus className="mr-1 h-4 w-4" /> {t("common:new")}
          </Button>
        </div>
      </div>

      <div className="flex items-center gap-2 flex-wrap">
        <div className="flex rounded-md border">
          {(["expense", "income", "transfer", "all"] as const).map((v) => (
            <Button
              key={v}
              size="sm"
              variant={typeFilter === v ? "default" : "ghost"}
              className="rounded-none first:rounded-l-md last:rounded-r-md"
              onClick={() => {
                setTypeFilter(v);
                setPage(0);
              }}
            >
              {v === "expense" ? t("transaction:filters.expenses") : v === "income" ? t("transaction:filters.income") : v === "transfer" ? t("transaction:filters.transfers") : t("transaction:filters.all")}
            </Button>
          ))}
        </div>
        <div className="h-6 w-px bg-border" />
        <Select
          value={categoryFilter}
          onValueChange={(v) => { setCategoryFilter(v); setPage(0); }}
        >
          <SelectTrigger className="w-[160px] h-8 text-sm">
            <SelectValue placeholder={t("transaction:category")} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">{t("transaction:filters.allCategories")}</SelectItem>
            {categories?.map((c) => (
              <SelectItem key={c.id} value={String(c.id)}>
                {c.icon} {getCategoryDisplayName(c, t)}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Select
          value={currencyFilter}
          onValueChange={(v) => { setCurrencyFilter(v); setPage(0); }}
        >
          <SelectTrigger className="w-[100px] h-8 text-sm">
            <SelectValue placeholder={t("transaction:currency")} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">{t("transaction:filters.allCurrencies")}</SelectItem>
            <SelectItem value="PLN">PLN</SelectItem>
            <SelectItem value="EUR">EUR</SelectItem>
            <SelectItem value="USD">USD</SelectItem>
            <SelectItem value="GBP">GBP</SelectItem>
          </SelectContent>
        </Select>
        {linkedBankAccounts && linkedBankAccounts.length > 0 && (
          <Select
            value={bankFilter}
            onValueChange={(v) => { setBankFilter(v); setPage(0); }}
          >
            <SelectTrigger className="w-[160px] h-8 text-sm">
              <SelectValue placeholder={t("transaction:filters.allBanks")} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">{t("transaction:filters.allBanks")}</SelectItem>
              {linkedBankAccounts.map((ba) => (
                <SelectItem key={ba.id} value={String(ba.id)}>
                  {ba.display_name} ({ba.currency})
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        )}
        <div className="h-6 w-px bg-border" />
        <div className="flex items-center gap-1">
          <Input
            type="text"
            inputMode="decimal"
            placeholder={t("transaction:filters.minAmount")}
            className="w-[90px] h-8 text-sm"
            value={amountMin}
            onChange={(e) => setAmountMin(e.target.value)}
            onBlur={() => { setAppliedAmountMin(amountMin); setPage(0); }}
            onKeyDown={(e: KeyboardEvent<HTMLInputElement>) => {
              if (e.key === "Enter") { setAppliedAmountMin(amountMin); setPage(0); }
            }}
          />
          <span className="text-muted-foreground text-sm">&ndash;</span>
          <Input
            type="text"
            inputMode="decimal"
            placeholder={t("transaction:filters.maxAmount")}
            className="w-[90px] h-8 text-sm"
            value={amountMax}
            onChange={(e) => setAmountMax(e.target.value)}
            onBlur={() => { setAppliedAmountMax(amountMax); setPage(0); }}
            onKeyDown={(e: KeyboardEvent<HTMLInputElement>) => {
              if (e.key === "Enter") { setAppliedAmountMax(amountMax); setPage(0); }
            }}
          />
        </div>
        <div className="h-6 w-px bg-border" />
        <DateRangePicker value={range} onChange={setRange} />
      </div>

      {isLoading ? (
        <div className="space-y-2">
          {[...Array(5)].map((_, i) => (
            <Skeleton key={i} className="h-10 w-full" />
          ))}
        </div>
      ) : !txs || txs.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <div className="rounded-full bg-muted p-3 mb-3">
            <Receipt className="h-5 w-5 text-muted-foreground" />
          </div>
          <p className="text-sm font-medium">{t("transaction:noTransactions")}</p>
          <p className="text-sm text-muted-foreground mt-1">{t("transaction:createToGetStarted")}</p>
        </div>
      ) : (
        <>
          <div className="rounded-lg border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-10">
                  <Checkbox
                    checked={allSelected ? true : someSelected ? "indeterminate" : false}
                    onCheckedChange={toggleSelectAll}
                  />
                </TableHead>
                <TableHead>
                  <button
                    className="inline-flex items-center gap-1 hover:text-foreground"
                    onClick={() => {
                      if (sortBy === "date") {
                        setSortOrder(sortOrder === "desc" ? "asc" : "desc");
                      } else {
                        setSortBy("date");
                        setSortOrder("desc");
                      }
                      setPage(0);
                    }}
                  >
                    {t("transaction:date")}
                    {sortBy === "date" && (sortOrder === "asc" ? <ArrowUp className="h-3 w-3" /> : <ArrowDown className="h-3 w-3" />)}
                  </button>
                </TableHead>
                <TableHead>
                  <button
                    className="inline-flex items-center gap-1 hover:text-foreground"
                    onClick={() => {
                      if (sortBy === "description") {
                        setSortOrder(sortOrder === "desc" ? "asc" : "desc");
                      } else {
                        setSortBy("description");
                        setSortOrder("desc");
                      }
                      setPage(0);
                    }}
                  >
                    {t("transaction:description")}
                    {sortBy === "description" && (sortOrder === "asc" ? <ArrowUp className="h-3 w-3" /> : <ArrowDown className="h-3 w-3" />)}
                  </button>
                </TableHead>
                <TableHead>{t("transaction:category")}</TableHead>
                <TableHead className="text-right">
                  <button
                    className="inline-flex items-center gap-1 hover:text-foreground ml-auto"
                    onClick={() => {
                      if (sortBy === "amount") {
                        setSortOrder(sortOrder === "desc" ? "asc" : "desc");
                      } else {
                        setSortBy("amount");
                        setSortOrder("desc");
                      }
                      setPage(0);
                    }}
                  >
                    {t("transaction:amount")}
                    {sortBy === "amount" && (sortOrder === "asc" ? <ArrowUp className="h-3 w-3" /> : <ArrowDown className="h-3 w-3" />)}
                  </button>
                </TableHead>
                <TableHead className="w-24" />
              </TableRow>
            </TableHeader>
            <TableBody>
              {txs.map((tx) => (
                <TableRow key={tx.id} className={`cursor-pointer ${selectedIds.has(tx.id) ? "bg-muted/50" : ""}`} onClick={() => setDetailTxId(tx.id)}>
                  <TableCell onClick={(e) => e.stopPropagation()}>
                    <Checkbox
                      checked={selectedIds.has(tx.id)}
                      onCheckedChange={() => toggleSelection(tx.id)}
                    />
                  </TableCell>
                  <TableCell className="text-muted-foreground">
                    {formatDate(tx.date)}
                  </TableCell>
                  <TableCell>
                    <span className="flex items-center gap-1.5">
                      {isMultiMember && (() => {
                        const member = memberMap.get(tx.created_by);
                        if (!member) return null;
                        return (
                          <MemberAvatar
                            displayName={member.display_name}
                            avatarUrl={member.avatar_url}
                            isCurrentUser={tx.created_by === currentUser?.id}
                          />
                        );
                      })()}
                      {tx.description}
                      {tx.type === "transfer" && (
                        <Badge variant="secondary" className="text-xs gap-1">
                          <ArrowLeftRight className="h-3 w-3" />
                          {t("transaction:transfer")}
                        </Badge>
                      )}
                      {tx.source === "bank" && (
                        <Badge variant="outline" className="text-xs gap-1">
                          <Building2 className="h-3 w-3" />
                          {tx.bank_name || t("workspace:bank.bankLabel")}
                        </Badge>
                      )}
                    </span>
                  </TableCell>
                  <TableCell className="text-sm text-muted-foreground">
                    {tx.entries.length > 0 && (() => {
                      const cat = categoryMap.get(tx.entries[0].category_id);
                      return cat ? `${cat.icon} ${getCategoryDisplayName(cat, t)}` : "";
                    })()}
                  </TableCell>
                  <TableCell className="text-right">
                    <MoneyDisplay
                      value={tx.total_amount}
                      className={tx.type === "income" ? "text-green-600" : undefined}
                    />
                  </TableCell>
                  <TableCell onClick={(e) => e.stopPropagation()}>
                    <div className="flex gap-1">
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => setEditTx(tx)}
                      >
                        <Pencil className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => setDeleteId(tx.id)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          </div>

          <div className="flex justify-center gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={page === 0}
              onClick={() => setPage((p) => p - 1)}
            >
              {t("common:previous")}
            </Button>
            <Button
              variant="outline"
              size="sm"
              disabled={!txs || txs.length < PAGE_SIZE}
              onClick={() => setPage((p) => p + 1)}
            >
              {t("common:next")}
            </Button>
          </div>
        </>
      )}

      {/* Create Dialog */}
      <CreateTransactionDialog open={createOpen} onOpenChange={setCreateOpen} />

      {/* Edit Dialog */}
      {editTx && (
        <EditTransactionDialog
          open={!!editTx}
          onOpenChange={(open) => !open && setEditTx(null)}
          transaction={editTx}
        />
      )}

      {/* Detail Dialog */}
      {detailTxId !== null && (
        <TransactionDetailDialog
          open={detailTxId !== null}
          onOpenChange={(open) => !open && setDetailTxId(null)}
          transactionId={detailTxId}
          onEdit={(tx) => { setDetailTxId(null); setEditTx(tx); }}
          onDelete={(id) => { setDetailTxId(null); setDeleteId(id); }}
        />
      )}

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => !open && setDeleteId(null)}
        title={t("transaction:deleteTitle")}
        description={t("transaction:deleteDescription")}
        onConfirm={handleDelete}
        loading={deleteMutation.isPending}
      />

      <BulkActionBar
        selectedIds={selectedIds}
        selectedTotals={selectedTotals}
        onClearSelection={clearSelection}
      />
    </div>
  );
}
