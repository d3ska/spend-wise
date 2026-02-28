import { useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { Building2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useLinkBankAccount } from "@/api/bank-accounts";
import { getErrorMessage } from "@/lib/errors";
import type { BankAccount } from "@/types";

interface AutoLinkDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  accounts: BankAccount[];
  workspaceId: number;
  workspaceName: string;
}

export function AutoLinkDialog({
  open,
  onOpenChange,
  accounts,
  workspaceId,
  workspaceName,
}: AutoLinkDialogProps) {
  const { t } = useTranslation(["workspace", "common"]);
  const [selected, setSelected] = useState<Set<number>>(
    () => new Set(accounts.map((a) => a.id)),
  );
  const [linking, setLinking] = useState(false);
  const linkAccount = useLinkBankAccount(workspaceId);

  const toggle = (id: number) => {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  const handleLink = async () => {
    setLinking(true);
    try {
      for (const id of selected) {
        await linkAccount.mutateAsync(id);
      }
      toast.success(
        t("workspace:bank.linkedCount", { count: selected.size, name: workspaceName }),
      );
      onOpenChange(false);
    } catch (err) {
      toast.error(getErrorMessage(err, t("workspace:bank.linkSomeFailed")));
    } finally {
      setLinking(false);
    }
  };

  if (accounts.length === 0) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t("workspace:bank.linkTitle")}</DialogTitle>
          <DialogDescription>
            {t("workspace:bank.foundAccounts")} {accounts.length} {t("workspace:bank.accountsFrom")}{" "}
            <strong>{workspaceName}</strong>.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-2 max-h-64 overflow-y-auto">
          {accounts.map((acct) => (
            <label
              key={acct.id}
              className="flex items-center gap-3 rounded-md border p-3 cursor-pointer hover:bg-muted/50"
            >
              <Checkbox
                checked={selected.has(acct.id)}
                onCheckedChange={() => toggle(acct.id)}
              />
              <Building2 className="h-4 w-4 text-muted-foreground shrink-0" />
              <div className="min-w-0">
                <p className="text-sm font-medium truncate">{acct.name}</p>
                <p className="text-xs text-muted-foreground truncate">
                  {acct.institution_name}
                  {acct.currency ? ` · ${acct.currency}` : ""}
                  {acct.iban ? ` · ${acct.iban}` : ""}
                </p>
              </div>
            </label>
          ))}
        </div>

        <DialogFooter className="gap-2 sm:gap-0">
          <Button
            variant="ghost"
            onClick={() => onOpenChange(false)}
            disabled={linking}
          >
            {t("common:skip")}
          </Button>
          <Button
            onClick={handleLink}
            disabled={selected.size === 0 || linking}
          >
            {linking
              ? t("workspace:bank.linking")
              : `${t("workspace:bank.linkTo")} ${workspaceName}`}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
