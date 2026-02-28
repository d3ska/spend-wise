import { useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { Building2, Search } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { useInstitutions, useInitiateConnection } from "@/api/bank-connections";
import { getErrorMessage } from "@/lib/errors";

interface Props {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

const COUNTRY_CODES = [
  { code: "PL", key: "poland" },
  { code: "DE", key: "germany" },
  { code: "GB", key: "unitedKingdom" },
  { code: "FR", key: "france" },
  { code: "ES", key: "spain" },
  { code: "IT", key: "italy" },
  { code: "NL", key: "netherlands" },
  { code: "AT", key: "austria" },
  { code: "BE", key: "belgium" },
  { code: "SE", key: "sweden" },
] as const;

export function ConnectBankDialog({ open, onOpenChange }: Props) {
  const { t } = useTranslation(["workspace", "transaction", "common"]);
  const [country, setCountry] = useState("PL");
  const [search, setSearch] = useState("");
  const { data: institutions, isLoading } = useInstitutions(country);
  const initiate = useInitiateConnection();

  const filtered = (institutions ?? []).filter((inst) =>
    inst.name.toLowerCase().includes(search.toLowerCase()),
  );

  function handleSelect(institutionId: string, institutionName: string) {
    // We don't know connection_id yet, so we use a placeholder redirect URL.
    // After initiate returns the connection_id, we'll set it in the actual redirect.
    const baseCallback = `${window.location.origin}/bank/callback`;
    initiate.mutate(
      {
        institution_id: institutionId,
        institution_name: institutionName,
        redirect_url: baseCallback,
        country,
      },
      {
        onSuccess: (result) => {
          window.location.href = result.auth_link;
        },
        onError: (err) => {
          toast.error(getErrorMessage(err, t("workspace:bank.connectionFailed")));
        },
      },
    );
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg max-h-[80vh] flex flex-col">
        <DialogHeader>
          <DialogTitle>{t("workspace:bank.connectTitle")}</DialogTitle>
        </DialogHeader>

        <div className="flex gap-2">
          <Select value={country} onValueChange={setCountry}>
            <SelectTrigger className="w-[160px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {COUNTRY_CODES.map((c) => (
                <SelectItem key={c.code} value={c.code}>
                  {t(`transaction:countries.${c.key}`)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          <div className="relative flex-1">
            <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder={t("workspace:bank.searchBanks")}
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="pl-8"
            />
          </div>
        </div>

        <div className="flex-1 overflow-y-auto space-y-1 min-h-[200px]">
          {isLoading ? (
            Array.from({ length: 5 }).map((_, i) => (
              <Skeleton key={i} className="h-12 w-full" />
            ))
          ) : filtered.length === 0 ? (
            <p className="text-sm text-muted-foreground py-8 text-center">
              {t("workspace:bank.noBanksFound")}
            </p>
          ) : (
            filtered.map((inst) => (
              <button
                key={inst.id}
                onClick={() => handleSelect(inst.id, inst.name)}
                disabled={initiate.isPending}
                className="w-full flex items-center gap-3 rounded-md px-3 py-2 text-left hover:bg-muted transition-colors disabled:opacity-50"
              >
                {inst.logo ? (
                  <img
                    src={inst.logo}
                    alt={inst.name}
                    className="h-8 w-8 rounded object-contain"
                  />
                ) : (
                  <Building2 className="h-8 w-8 text-muted-foreground" />
                )}
                <span className="text-sm font-medium">{inst.name}</span>
              </button>
            ))
          )}
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={initiate.isPending}
          >
            {t("common:cancel")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
