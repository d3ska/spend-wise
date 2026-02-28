import { useEffect, useRef } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { Skeleton } from "@/components/ui/skeleton";
import { useCompleteConnection } from "@/api/bank-connections";

export default function BankCallbackPage() {
  const { t } = useTranslation(["workspace"]);
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const complete = useCompleteConnection();
  const navigatedRef = useRef(false);
  const calledRef = useRef(false);

  const code = searchParams.get("code");

  useEffect(() => {
    if (!code || calledRef.current) return;
    calledRef.current = true;

    complete.mutateAsync({ code }).then(
      (accounts) => {
        if (navigatedRef.current) return;
        navigatedRef.current = true;
        toast.success(t("workspace:bank.connected"));
        navigate("/settings", {
          replace: true,
          state: { newAccounts: accounts },
        });
      },
      () => {
        if (navigatedRef.current) return;
        navigatedRef.current = true;
        toast.error(t("workspace:bank.connectionFailed"));
        navigate("/settings", { replace: true });
      },
    );

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  if (!code) {
    return (
      <div className="flex min-h-screen items-center justify-center p-4">
        <div className="text-center space-y-4">
          <p className="text-destructive">
            {t("workspace:bank.missingParams")}
          </p>
          <a href="/settings" className="text-sm underline">
            {t("workspace:bank.backToSettings")}
          </a>
        </div>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="space-y-4 w-80 text-center">
        <Skeleton className="h-8 w-48 mx-auto" />
        <p className="text-sm text-muted-foreground">
          {t("workspace:bank.connecting")}
        </p>
      </div>
    </div>
  );
}
