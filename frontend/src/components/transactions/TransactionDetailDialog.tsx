import { useTranslation } from "react-i18next";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useTransaction } from "@/api/transactions";
import { useLinkedBankAccounts } from "@/api/bank-accounts";
import { useWorkspaceContext } from "@/hooks/useWorkspace";
import { useCategories } from "@/api/categories";
import { MoneyDisplay } from "@/components/shared/MoneyDisplay";
import { formatDate } from "@/lib/dates";
import { getCategoryDisplayName } from "@/lib/categoryUtils";
import type { Transaction } from "@/types";
import { ArrowLeftRight, Building2, Pencil, Trash2 } from "lucide-react";

interface Props {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  transactionId: number;
  onEdit: (tx: Transaction) => void;
  onDelete: (id: number) => void;
}

function sourceLabel(
  source: string,
  t: (key: string) => string,
): string {
  switch (source) {
    case "manual":
      return t("transaction:detail.sourceManual");
    case "import":
      return t("transaction:detail.sourceImport");
    case "bank":
      return t("transaction:detail.sourceBank");
    default:
      return source;
  }
}

function typeLabel(
  type: string,
  t: (key: string) => string,
): string {
  switch (type) {
    case "expense":
      return t("transaction:expense");
    case "income":
      return t("transaction:income");
    case "transfer":
      return t("transaction:transfer");
    default:
      return type;
  }
}

export function TransactionDetailDialog({
  open,
  onOpenChange,
  transactionId,
  onEdit,
  onDelete,
}: Props) {
  const { t } = useTranslation(["transaction", "common"]);
  const { workspaceId } = useWorkspaceContext();
  const { data: tx, isLoading } = useTransaction(workspaceId, transactionId);
  const { data: bankAccounts } = useLinkedBankAccounts(workspaceId);
  const { data: categories } = useCategories(workspaceId);

  const bankAccount = bankAccounts?.find(
    (ba) => ba.id === tx?.bank_account_id,
  );

  const entry = tx?.entries[0];
  const category = categories?.find((c) => c.id === entry?.category_id);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg max-h-[90vh] overflow-y-auto">
        {isLoading || !tx ? (
          <div className="flex items-center justify-center py-12">
            <div className="text-sm text-muted-foreground">
              {t("common:loading")}
            </div>
          </div>
        ) : (
          <>
            <DialogHeader>
              <div className="flex items-center gap-2 flex-wrap">
                <DialogTitle className="text-lg">
                  {tx.description}
                </DialogTitle>
                {tx.type === "transfer" && (
                  <Badge variant="outline" className="gap-1">
                    <ArrowLeftRight className="size-3" />
                    {typeLabel(tx.type, t)}
                  </Badge>
                )}
                <Badge variant="secondary">
                  {sourceLabel(tx.source, t)}
                </Badge>
              </div>
            </DialogHeader>

            <div className="space-y-6">
              {/* Amount and date */}
              <div className="flex items-baseline justify-between">
                <MoneyDisplay
                  value={tx.total_amount}
                  className="text-2xl font-semibold"
                />
                <span className="text-sm text-muted-foreground">
                  {formatDate(tx.date)}
                </span>
              </div>

              {/* Type (for non-transfer) */}
              {tx.type !== "transfer" && (
                <div className="grid grid-cols-[140px_1fr] gap-y-2 text-sm">
                  <span className="text-muted-foreground">
                    {t("transaction:type")}
                  </span>
                  <span>{typeLabel(tx.type, t)}</span>
                </div>
              )}

              {/* Category */}
              {category && (
                <div className="grid grid-cols-[140px_1fr] gap-y-2 text-sm">
                  <span className="text-muted-foreground">
                    {t("transaction:category")}
                  </span>
                  <span>
                    {category.icon} {getCategoryDisplayName(category, t)}
                  </span>
                </div>
              )}

              {/* Banking details */}
              {tx.source === "bank" && (
                <div className="space-y-3">
                  <div className="flex items-center gap-2 text-sm font-medium">
                    <Building2 className="size-4 text-muted-foreground" />
                    {t("transaction:detail.bankingDetails")}
                  </div>
                  <div className="grid grid-cols-[140px_1fr] gap-y-2 text-sm">
                    {tx.bank_name && (
                      <>
                        <span className="text-muted-foreground">
                          {t("transaction:detail.source")}
                        </span>
                        <span>{tx.bank_name}</span>
                      </>
                    )}
                    {bankAccount && (
                      <>
                        <span className="text-muted-foreground">
                          {t("transaction:detail.bankAccount")}
                        </span>
                        <span>{bankAccount.display_name}</span>
                      </>
                    )}
                    {tx.iban && (
                      <>
                        <span className="text-muted-foreground">
                          {t("transaction:detail.iban")}
                        </span>
                        <span className="font-mono text-xs">
                          {tx.iban}
                        </span>
                      </>
                    )}
                    {tx.counterparty_iban && (
                      <>
                        <span className="text-muted-foreground">
                          {t("transaction:detail.counterpartyIban")}
                        </span>
                        <span className="font-mono text-xs">
                          {tx.counterparty_iban}
                        </span>
                      </>
                    )}
                  </div>
                </div>
              )}

              {/* Notes */}
              {tx.notes && (
                <div className="space-y-1">
                  <span className="text-sm text-muted-foreground">
                    {t("transaction:notes")}
                  </span>
                  <p className="text-sm whitespace-pre-wrap">{tx.notes}</p>
                </div>
              )}

              {/* Metadata */}
              <div className="border-t pt-4">
                <div className="grid grid-cols-[140px_1fr] gap-y-2 text-xs text-muted-foreground">
                  <span>{t("transaction:detail.createdAt")}</span>
                  <span>{formatDate(tx.created_at)}</span>
                  <span>{t("transaction:detail.updatedAt")}</span>
                  <span>{formatDate(tx.updated_at)}</span>
                </div>
              </div>

              {/* Actions */}
              <div className="flex justify-end gap-2 border-t pt-4">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => {
                    onEdit(tx);
                    onOpenChange(false);
                  }}
                >
                  <Pencil className="size-4 mr-1" />
                  {t("common:edit")}
                </Button>
                <Button
                  variant="destructive"
                  size="sm"
                  onClick={() => {
                    onDelete(tx.id);
                    onOpenChange(false);
                  }}
                >
                  <Trash2 className="size-4 mr-1" />
                  {t("common:delete")}
                </Button>
              </div>
            </div>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}
