import { useState, useMemo } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { Plus, Pencil, Sparkles, Trash2, Wand2, ChevronRight, ChevronDown, CircleHelp } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
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
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import {
  Popover,
  PopoverTrigger,
  PopoverContent,
} from "@/components/ui/popover";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Skeleton } from "@/components/ui/skeleton";
import { Badge } from "@/components/ui/badge";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog";
import {
  useRules,
  useCreateRule,
  useUpdateRule,
  useDeleteRule,
  useToggleRule,
} from "@/api/rules";
import { useCategories } from "@/api/categories";
import { useLinkedBankAccounts } from "@/api/bank-accounts";
import { useApplyRules } from "@/api/transactions";
import { useWorkspaceContext } from "@/hooks/useWorkspace";
import { getErrorMessage } from "@/lib/errors";
import { getCategoryDisplayName } from "@/lib/categoryUtils";
import type { Rule } from "@/types";

type MatchMode = "contains" | "starts_with" | "exact" | "advanced";

const PRIORITY_LEVELS = [
  { value: "0", labelKey: "transaction:rules.low" },
  { value: "10", labelKey: "transaction:rules.normal" },
  { value: "20", labelKey: "transaction:rules.high" },
  { value: "30", labelKey: "transaction:rules.critical" },
] as const;

function escapeRegex(s: string): string {
  return s.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

function buildPattern(
  mode: MatchMode,
  keywords: string,
  matchText: string,
  caseSensitive: boolean,
  rawPattern: string,
): string {
  if (mode === "advanced") return rawPattern;
  const prefix = caseSensitive ? "" : "(?i)";
  if (mode === "contains") {
    const parts = keywords
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean)
      .map(escapeRegex);
    if (parts.length === 0) return "";
    return `${prefix}${parts.join("|")}`;
  }
  const escaped = escapeRegex(matchText.trim());
  if (!escaped) return "";
  if (mode === "starts_with") return `${prefix}^${escaped}`;
  // exact
  return `${prefix}^${escaped}$`;
}

function parsePattern(raw: string): {
  mode: MatchMode;
  keywords: string;
  matchText: string;
  caseSensitive: boolean;
} {
  let p = raw;
  let caseSensitive = true;
  if (p.startsWith("(?i)")) {
    caseSensitive = false;
    p = p.slice(4);
  }

  // Exact: ^text$
  if (p.startsWith("^") && p.endsWith("$") && !p.slice(1, -1).includes("|")) {
    return { mode: "exact", keywords: "", matchText: p.slice(1, -1), caseSensitive };
  }
  // Starts with: ^text (no pipe)
  if (p.startsWith("^") && !p.includes("|")) {
    return { mode: "starts_with", keywords: "", matchText: p.slice(1), caseSensitive };
  }
  // Contains keywords: word1|word2|word3 (no anchors, no groups)
  if (!p.includes("^") && !p.includes("$") && !p.includes("(") && !p.includes("[")) {
    const parts = p.split("|");
    if (parts.length > 0 && parts.every((s) => s.length > 0)) {
      return { mode: "contains", keywords: parts.join(", "), matchText: "", caseSensitive };
    }
  }
  // Fallback
  return { mode: "advanced", keywords: "", matchText: "", caseSensitive: true };
}

export default function RulesPage() {
  const { t } = useTranslation(["transaction", "common"]);
  const { workspaceId } = useWorkspaceContext();
  const { data: rules, isLoading } = useRules(workspaceId);
  const { data: categories } = useCategories(workspaceId);
  const { data: linkedBankAccounts } = useLinkedBankAccounts(workspaceId);
  const createMutation = useCreateRule(workspaceId);
  const updateMutation = useUpdateRule(workspaceId);
  const deleteMutation = useDeleteRule(workspaceId);
  const toggleMutation = useToggleRule(workspaceId);
  const applyRules = useApplyRules(workspaceId);

  const [dialogOpen, setDialogOpen] = useState(false);
  const [editing, setEditing] = useState<Rule | null>(null);
  const [deleteId, setDeleteId] = useState<number | null>(null);
  const [matchMode, setMatchMode] = useState<MatchMode>("contains");
  const [keywords, setKeywords] = useState("");
  const [matchText, setMatchText] = useState("");
  const [caseSensitive, setCaseSensitive] = useState(false);
  const [rawPattern, setRawPattern] = useState("");
  const [categoryId, setCategoryId] = useState("");
  const [priority, setPriority] = useState("10");
  const [amountMin, setAmountMin] = useState("");
  const [amountMax, setAmountMax] = useState("");
  const [bankAccountId, setBankAccountId] = useState<string>("any");
  const [counterpartyIban, setCounterpartyIban] = useState("");
  const [filtersOpen, setFiltersOpen] = useState(false);

  const activeFilterCount = useMemo(() => {
    let count = 0;
    if (priority !== "10") count++;
    if (amountMin.trim() || amountMax.trim()) count++;
    if (bankAccountId !== "any") count++;
    if (counterpartyIban.trim()) count++;
    return count;
  }, [priority, amountMin, amountMax, bankAccountId, counterpartyIban]);

  const computedPattern = useMemo(
    () => buildPattern(matchMode, keywords, matchText, caseSensitive, rawPattern),
    [matchMode, keywords, matchText, caseSensitive, rawPattern],
  );

  const priorityLabel = (p: number): string => {
    const level = PRIORITY_LEVELS.find((l) => Number(l.value) === p);
    return level ? t(level.labelKey) : `Custom (${p})`;
  };

  const getCategoryName = (id: number) => {
    const cat = categories?.find((c) => c.id === id);
    if (!cat) return `#${id}`;
    return getCategoryDisplayName(cat, t);
  };

  const resetPatternFields = () => {
    setMatchMode("contains");
    setKeywords("");
    setMatchText("");
    setCaseSensitive(false);
    setRawPattern("");
  };

  const openCreate = () => {
    setEditing(null);
    resetPatternFields();
    setCategoryId("");
    setPriority("10");
    setAmountMin("");
    setAmountMax("");
    setBankAccountId("any");
    setCounterpartyIban("");
    setFiltersOpen(false);
    setDialogOpen(true);
  };

  const openEdit = (rule: Rule) => {
    setEditing(rule);
    const parsed = parsePattern(rule.match_pattern);
    setMatchMode(parsed.mode);
    setKeywords(parsed.keywords);
    setMatchText(parsed.matchText);
    setCaseSensitive(parsed.caseSensitive);
    setRawPattern(parsed.mode === "advanced" ? rule.match_pattern : "");
    setCategoryId(String(rule.target_category_id));
    setAmountMin(rule.amount_min ?? "");
    setAmountMax(rule.amount_max ?? "");
    setBankAccountId(rule.bank_account_id != null ? String(rule.bank_account_id) : "any");
    setCounterpartyIban(rule.counterparty_iban ?? "");
    // Snap to nearest predefined level
    const nearest = PRIORITY_LEVELS.reduce((prev, curr) =>
      Math.abs(Number(curr.value) - rule.priority) < Math.abs(Number(prev.value) - rule.priority) ? curr : prev,
    );
    setPriority(nearest.value);
    const hasOptionalFilters =
      nearest.value !== "10" ||
      (rule.amount_min != null && rule.amount_min !== "") ||
      (rule.amount_max != null && rule.amount_max !== "") ||
      rule.bank_account_id != null ||
      (rule.counterparty_iban != null && rule.counterparty_iban !== "");
    setFiltersOpen(hasOptionalFilters);
    setDialogOpen(true);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const body = {
      match_pattern: computedPattern,
      target_category_id: parseInt(categoryId),
      priority: parseInt(priority),
      amount_min: amountMin.trim() || null,
      amount_max: amountMax.trim() || null,
      bank_account_id: bankAccountId !== "any" ? parseInt(bankAccountId) : null,
      counterparty_iban: counterpartyIban.trim() || null,
    };
    try {
      if (editing) {
        await updateMutation.mutateAsync({ id: editing.id, ...body });
        toast.success(t("transaction:rules.updated"));
      } else {
        await createMutation.mutateAsync(body);
        toast.success(t("transaction:rules.created"));
      }
      setDialogOpen(false);
    } catch (err) {
      toast.error(getErrorMessage(err, t("transaction:rules.saveFailed")));
    }
  };

  const handleToggle = async (rule: Rule) => {
    try {
      await toggleMutation.mutateAsync({
        id: rule.id,
        enabled: !rule.enabled,
      });
    } catch (err) {
      toast.error(getErrorMessage(err, t("transaction:rules.toggleFailed")));
    }
  };

  const handleDelete = async () => {
    if (deleteId === null) return;
    try {
      await deleteMutation.mutateAsync(deleteId);
      toast.success(t("transaction:rules.deleted"));
      setDeleteId(null);
    } catch (err) {
      toast.error(getErrorMessage(err, t("transaction:rules.deleteFailed")));
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">{t("transaction:rules.title")}</h1>
          <p className="text-sm text-muted-foreground">{t("transaction:rules.subtitle")}</p>
        </div>
        <div className="flex gap-2">
          <Button
            size="sm"
            variant="outline"
            onClick={() =>
              applyRules.mutate(undefined, {
                onSuccess: (data) =>
                  toast.success(
                    t("transaction:rules.applied", { updated: data.updated, total: data.total }),
                  ),
                onError: (err) => toast.error(getErrorMessage(err, t("transaction:rules.applyFailed"))),
              })
            }
            disabled={applyRules.isPending}
          >
            <Wand2 className="mr-1 h-4 w-4" />
            {applyRules.isPending ? t("transaction:rules.applying") : t("transaction:rules.applyRules")}
          </Button>
          <Button size="sm" onClick={openCreate}>
            <Plus className="mr-1 h-4 w-4" /> {t("common:new")}
          </Button>
        </div>
      </div>

      {isLoading ? (
        <div className="space-y-2">
          {[...Array(3)].map((_, i) => (
            <Skeleton key={i} className="h-10 w-full" />
          ))}
        </div>
      ) : !rules || rules.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <div className="rounded-full bg-muted p-3 mb-3">
            <Sparkles className="h-5 w-5 text-muted-foreground" />
          </div>
          <p className="text-sm font-medium">{t("transaction:rules.noRulesYet")}</p>
          <p className="text-sm text-muted-foreground mt-1">{t("transaction:rules.createToGetStarted")}</p>
        </div>
      ) : (
        <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t("transaction:rules.keywords")}</TableHead>
              <TableHead>{t("transaction:category")}</TableHead>
              <TableHead>{t("transaction:rules.filters_column")}</TableHead>
              <TableHead>{t("transaction:rules.priority")}</TableHead>
              <TableHead>{t("transaction:rules.status")}</TableHead>
              <TableHead className="w-24" />
            </TableRow>
          </TableHeader>
          <TableBody>
            {rules.map((rule) => (
              <TableRow key={rule.id}>
                <TableCell className="text-sm">
                  {(() => {
                    const parsed = parsePattern(rule.match_pattern);
                    if (parsed.mode === "contains") {
                      const all = parsed.keywords.split(",").map((kw) => kw.trim()).filter(Boolean);
                      const visible = all.slice(0, 6);
                      const remaining = all.length - visible.length;
                      return (
                        <div className="flex flex-wrap gap-1 items-center">
                          {visible.map((kw, i) => (
                            <Badge key={i} variant="secondary" className="font-normal">
                              {kw}
                            </Badge>
                          ))}
                          {remaining > 0 && (
                            <span className="text-xs text-muted-foreground">{t("transaction:rules.moreCategories", { count: remaining })}</span>
                          )}
                        </div>
                      );
                    }
                    if (parsed.mode === "starts_with") {
                      return <span>{t("transaction:rules.startsWith")} <Badge variant="secondary" className="font-normal">{parsed.matchText}</Badge></span>;
                    }
                    if (parsed.mode === "exact") {
                      return <span>{t("transaction:rules.exactMatch")} <Badge variant="secondary" className="font-normal">{parsed.matchText}</Badge></span>;
                    }
                    return <span className="font-mono">{rule.match_pattern}</span>;
                  })()}
                </TableCell>
                <TableCell>
                  {getCategoryName(rule.target_category_id)}
                </TableCell>
                <TableCell className="text-sm">
                  {(() => {
                    const badges: { key: string; label: string }[] = [];
                    if (rule.amount_min || rule.amount_max) {
                      badges.push({ key: "amt", label: `${rule.amount_min ?? "0"} – ${rule.amount_max ?? "∞"}` });
                    }
                    if (rule.bank_account_id != null) {
                      const ba = linkedBankAccounts?.find((b) => b.id === rule.bank_account_id);
                      badges.push({ key: "bank", label: ba ? ba.display_name : `#${rule.bank_account_id}` });
                    }
                    if (rule.counterparty_iban) {
                      const iban = rule.counterparty_iban;
                      badges.push({ key: "iban", label: `${iban.slice(0, 2)}...${iban.slice(-4)}` });
                    }
                    if (badges.length === 0) {
                      return <span className="text-muted-foreground">{t("transaction:rules.filtersNone")}</span>;
                    }
                    return (
                      <div className="flex flex-wrap gap-1">
                        {badges.map((b) => (
                          <Badge key={b.key} variant="outline" className="font-normal text-xs">
                            {b.label}
                          </Badge>
                        ))}
                      </div>
                    );
                  })()}
                </TableCell>
                <TableCell>{priorityLabel(rule.priority)}</TableCell>
                <TableCell>
                  <Switch
                    checked={rule.enabled}
                    onCheckedChange={() => handleToggle(rule)}
                  />
                </TableCell>
                <TableCell>
                  <div className="flex gap-1">
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => openEdit(rule)}
                    >
                      <Pencil className="h-4 w-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => setDeleteId(rule.id)}
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
      )}

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editing ? t("transaction:rules.editTitle") : t("transaction:rules.newTitle")}</DialogTitle>
            <DialogDescription>{t("transaction:rules.dialogDescription")}</DialogDescription>
          </DialogHeader>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-1">
              <Label>{t("transaction:rules.matchMode")}</Label>
              <Select value={matchMode} onValueChange={(v) => setMatchMode(v as MatchMode)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="contains">{t("transaction:rules.containsKeywords")}</SelectItem>
                  <SelectItem value="starts_with">{t("transaction:rules.startsWith")}</SelectItem>
                  <SelectItem value="exact">{t("transaction:rules.exactMatch")}</SelectItem>
                  <SelectItem value="advanced">{t("transaction:rules.advancedRegex")}</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {matchMode === "contains" && (
              <div className="space-y-1">
                <Label>{t("transaction:rules.keywords")}</Label>
                <Input
                  value={keywords}
                  onChange={(e) => setKeywords(e.target.value)}
                  placeholder={t("transaction:categories.keywordsPlaceholder")}
                  required
                />
                <p className="text-xs text-muted-foreground">
                  {t("transaction:rules.keywordsPlaceholder")}
                </p>
              </div>
            )}

            {(matchMode === "starts_with" || matchMode === "exact") && (
              <div className="space-y-1">
                <Label>{matchMode === "starts_with" ? t("transaction:rules.prefix") : t("transaction:rules.text")}</Label>
                <Input
                  value={matchText}
                  onChange={(e) => setMatchText(e.target.value)}
                  placeholder={matchMode === "starts_with" ? t("transaction:categories.startsWithPlaceholder") : t("transaction:categories.exactPlaceholder")}
                  required
                />
              </div>
            )}

            {matchMode === "advanced" && (
              <div className="space-y-1">
                <Label>{t("transaction:rules.regexPattern")}</Label>
                <Input
                  value={rawPattern}
                  onChange={(e) => setRawPattern(e.target.value)}
                  placeholder={t("transaction:categories.regexPlaceholder")}
                  className="font-mono"
                  required
                />
              </div>
            )}

            {matchMode !== "advanced" && (
              <div className="flex items-center gap-2">
                <Switch
                  id="case-sensitive"
                  checked={caseSensitive}
                  onCheckedChange={setCaseSensitive}
                />
                <Label htmlFor="case-sensitive" className="text-sm font-normal">
                  {t("transaction:rules.caseSensitive")}
                </Label>
              </div>
            )}

            {computedPattern && (
              <div className="rounded-md bg-muted px-3 py-2">
                <p className="text-xs text-muted-foreground mb-1">{t("transaction:rules.generatedPattern")}</p>
                <p className="font-mono text-sm break-all">{computedPattern}</p>
              </div>
            )}

            <div className="space-y-1">
              <Label>{t("transaction:rules.targetCategory")}</Label>
              <Select value={categoryId} onValueChange={setCategoryId}>
                <SelectTrigger>
                  <SelectValue placeholder={t("transaction:selectCategory")} />
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
            <div>
              <button
                type="button"
                className="flex items-center gap-1.5 text-sm font-medium hover:text-foreground text-muted-foreground transition-colors"
                onClick={() => setFiltersOpen((v) => !v)}
              >
                {filtersOpen ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
                {t("transaction:rules.optionalFilters")}
                {activeFilterCount > 0 && (
                  <Badge variant="secondary" className="ml-1 text-xs px-1.5 py-0">
                    {activeFilterCount}
                  </Badge>
                )}
              </button>

              {filtersOpen && (
                <div className="space-y-4 mt-3 pl-1 border-l-2 border-muted ml-1.5">
                  <div className="space-y-1 pl-3">
                    <Label>{t("transaction:rules.priority")}</Label>
                    <Select value={priority} onValueChange={setPriority}>
                      <SelectTrigger>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {PRIORITY_LEVELS.map((level) => (
                          <SelectItem key={level.value} value={level.value}>
                            {t(level.labelKey)}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <p className="text-xs text-muted-foreground">
                      {t("transaction:rules.priorityHelp")}
                    </p>
                  </div>
                  <div className="space-y-1 pl-3">
                    <Label>{t("transaction:rules.amountRange")}</Label>
                    <div className="flex gap-2">
                      <Input
                        type="number"
                        min="0"
                        step="0.01"
                        value={amountMin}
                        onChange={(e) => setAmountMin(e.target.value)}
                        placeholder={t("transaction:rules.amountMin")}
                      />
                      <Input
                        type="number"
                        min="0"
                        step="0.01"
                        value={amountMax}
                        onChange={(e) => setAmountMax(e.target.value)}
                        placeholder={t("transaction:rules.amountMax")}
                      />
                    </div>
                    <p className="text-xs text-muted-foreground">
                      {t("transaction:rules.amountRangeHelp")}
                    </p>
                  </div>
                  {linkedBankAccounts && linkedBankAccounts.length > 0 && (
                    <div className="space-y-1 pl-3">
                      <div className="flex items-center gap-1.5">
                        <Label>{t("transaction:rules.bankAccount")}</Label>
                        <Popover>
                          <PopoverTrigger asChild>
                            <button type="button" className="text-muted-foreground hover:text-foreground transition-colors">
                              <CircleHelp className="h-3.5 w-3.5" />
                            </button>
                          </PopoverTrigger>
                          <PopoverContent className="w-72 text-sm" side="top">
                            <p className="font-medium mb-1">{t("transaction:rules.bankAccountInfoTitle")}</p>
                            <p className="text-muted-foreground">{t("transaction:rules.bankAccountInfoBody")}</p>
                            <p className="text-muted-foreground italic mt-2">{t("transaction:rules.bankAccountInfoExample")}</p>
                          </PopoverContent>
                        </Popover>
                      </div>
                      <Select value={bankAccountId} onValueChange={setBankAccountId}>
                        <SelectTrigger>
                          <SelectValue placeholder={t("transaction:rules.anyAccount")} />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="any">{t("transaction:rules.anyAccount")}</SelectItem>
                          {linkedBankAccounts.map((ba) => (
                            <SelectItem key={ba.id} value={String(ba.id)}>
                              {ba.display_name} ({ba.currency})
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    </div>
                  )}
                  <div className="space-y-1 pl-3">
                    <div className="flex items-center gap-1.5">
                      <Label>{t("transaction:rules.counterpartyIban")}</Label>
                      <Popover>
                        <PopoverTrigger asChild>
                          <button type="button" className="text-muted-foreground hover:text-foreground transition-colors">
                            <CircleHelp className="h-3.5 w-3.5" />
                          </button>
                        </PopoverTrigger>
                        <PopoverContent className="w-72 text-sm" side="top">
                          <p className="font-medium mb-1">{t("transaction:rules.counterpartyIbanInfoTitle")}</p>
                          <p className="text-muted-foreground">{t("transaction:rules.counterpartyIbanInfoBody")}</p>
                          <p className="text-muted-foreground italic mt-2">{t("transaction:rules.counterpartyIbanInfoExample")}</p>
                        </PopoverContent>
                      </Popover>
                    </div>
                    <Input
                      value={counterpartyIban}
                      onChange={(e) => setCounterpartyIban(e.target.value)}
                      placeholder={t("transaction:rules.counterpartyIbanPlaceholder")}
                    />
                  </div>
                </div>
              )}
            </div>
            <DialogFooter>
              <Button
                type="submit"
                disabled={
                  createMutation.isPending || updateMutation.isPending || !computedPattern
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
        title={t("transaction:rules.deleteTitle")}
        description={t("transaction:rules.deleteDescription")}
        onConfirm={handleDelete}
        loading={deleteMutation.isPending}
      />
    </div>
  );
}
