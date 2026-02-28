import { useEffect, useRef } from "react";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { handleCallback } from "@/api/auth";
import { Skeleton } from "@/components/ui/skeleton";

export default function AuthCallbackPage() {
  const { t } = useTranslation(["auth"]);
  const { provider } = useParams<{ provider: string }>();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const qc = useQueryClient();
  const calledRef = useRef(false);

  const code = searchParams.get("code");
  const state = searchParams.get("state");

  const mutation = useMutation({
    mutationFn: async () => {
      if (!provider || !code || !state) {
        throw new Error(t("auth:missingParams"));
      }
      return handleCallback(provider, code, state);
    },
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: ["auth", "me"] });
      // Check for pending invite — redirect to invite page instead of dashboard
      const pendingInvite = localStorage.getItem("sw_pending_invite");
      if (pendingInvite) {
        navigate(`/invite/${pendingInvite}`, { replace: true });
      } else {
        navigate("/dashboard", { replace: true });
      }
    },
  });

  useEffect(() => {
    if (calledRef.current) return;
    calledRef.current = true;
    mutation.mutate();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  if (!provider || !code || !state) {
    return (
      <div className="flex min-h-screen items-center justify-center p-4">
        <div className="text-center space-y-4">
          <p className="text-destructive">
            {t("auth:missingParams")}
          </p>
          <a href="/login" className="text-sm underline">
            {t("auth:backToLogin")}
          </a>
        </div>
      </div>
    );
  }

  if (mutation.isError) {
    return (
      <div className="flex min-h-screen items-center justify-center p-4">
        <div className="text-center space-y-4">
          <p className="text-destructive">
            {t("auth:authFailed")}
          </p>
          <a href="/login" className="text-sm underline">
            {t("auth:backToLogin")}
          </a>
        </div>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="space-y-4 w-80">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-4 w-full" />
      </div>
    </div>
  );
}
