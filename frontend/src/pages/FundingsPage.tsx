import { useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { Plus, Wallet } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Skeleton } from "@/components/ui/skeleton";
import { MoneyDisplay } from "@/components/shared/MoneyDisplay";
import { useFundings, useRecordFunding } from "@/api/fundings";
import { useWorkspaceContext } from "@/hooks/useWorkspace";
import { getErrorMessage } from "@/lib/errors";

export default function FundingsPage() {
  const { t } = useTranslation(["transaction", "common"]);
  const { workspaceId } = useWorkspaceContext();
  const { data: fundings, isLoading } = useFundings(workspaceId);
  const recordMutation = useRecordFunding(workspaceId);

  const [dialogOpen, setDialogOpen] = useState(false);
  const [userId, setUserId] = useState("");
  const [yearMonth, setYearMonth] = useState("");
  const [amount, setAmount] = useState("");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await recordMutation.mutateAsync({
        user_id: parseInt(userId),
        year_month: yearMonth,
        amount: {
          amount: parseFloat(amount).toFixed(2),
          currency: "PLN",
        },
      });
      toast.success(t("transaction:fundings.recorded"));
      setDialogOpen(false);
      setUserId("");
      setYearMonth("");
      setAmount("");
    } catch (err) {
      toast.error(getErrorMessage(err, t("transaction:fundings.recordFailed")));
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">{t("transaction:fundings.title")}</h1>
          <p className="text-sm text-muted-foreground">{t("transaction:fundings.subtitle")}</p>
        </div>
        <Button size="sm" onClick={() => setDialogOpen(true)}>
          <Plus className="mr-1 h-4 w-4" /> {t("transaction:fundings.record")}
        </Button>
      </div>

      {isLoading ? (
        <div className="space-y-2">
          {[...Array(3)].map((_, i) => (
            <Skeleton key={i} className="h-10 w-full" />
          ))}
        </div>
      ) : !fundings || fundings.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <div className="rounded-full bg-muted p-3 mb-3">
            <Wallet className="h-5 w-5 text-muted-foreground" />
          </div>
          <p className="text-sm font-medium">{t("transaction:fundings.noFundings")}</p>
          <p className="text-sm text-muted-foreground mt-1">{t("transaction:fundings.createToGetStarted")}</p>
        </div>
      ) : (
        <div className="rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t("transaction:fundings.period")}</TableHead>
              <TableHead>{t("transaction:fundings.userId")}</TableHead>
              <TableHead className="text-right">{t("transaction:amount")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {fundings.map((f) => (
              <TableRow key={f.id}>
                <TableCell>{f.year_month}</TableCell>
                <TableCell>{f.user_id}</TableCell>
                <TableCell className="text-right">
                  <MoneyDisplay value={f.amount} />
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
            <DialogTitle>{t("transaction:fundings.recordTitle")}</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-1">
              <Label>{t("transaction:fundings.userId")}</Label>
              <Input
                type="number"
                value={userId}
                onChange={(e) => setUserId(e.target.value)}
                required
              />
            </div>
            <div className="space-y-1">
              <Label>{t("transaction:fundings.periodPlaceholder")}</Label>
              <Input
                type="month"
                value={yearMonth}
                onChange={(e) => setYearMonth(e.target.value)}
                required
              />
            </div>
            <div className="space-y-1">
              <Label>{t("transaction:amount")}</Label>
              <Input
                type="number"
                step="0.01"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                required
              />
            </div>
            <DialogFooter>
              <Button type="submit" disabled={recordMutation.isPending}>
                {recordMutation.isPending ? t("transaction:fundings.recording") : t("transaction:fundings.record")}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
