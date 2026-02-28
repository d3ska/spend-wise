import { useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { useAuth } from "@/hooks/useAuth";
import { usePreviewInvite, useAcceptInvite } from "@/api/invites";
import { getAuthURL } from "@/api/auth";
import { getErrorMessage } from "@/lib/errors";

const PENDING_INVITE_KEY = "sw_pending_invite";

export default function InvitePage() {
  const { t } = useTranslation(["workspace", "auth"]);
  const { code } = useParams<{ code: string }>();
  const navigate = useNavigate();
  const { user, isLoading: authLoading } = useAuth();
  const { data: preview, isLoading, isError } = usePreviewInvite(code ?? "");
  const acceptInvite = useAcceptInvite();

  // Auto-accept if user arrived back from SSO with a pending invite
  useEffect(() => {
    if (!user || !code) return;
    const pending = localStorage.getItem(PENDING_INVITE_KEY);
    if (pending === code) {
      localStorage.removeItem(PENDING_INVITE_KEY);
      handleAccept();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user, code]);

  const handleAccept = async () => {
    if (!code) return;
    try {
      const ws = await acceptInvite.mutateAsync(code);
      localStorage.setItem("sw_workspace_id", String(ws.id));
      toast.success(`${t("workspace:invite.joined")} ${ws.name}!`);
      navigate("/dashboard", { replace: true });
    } catch (err) {
      toast.error(getErrorMessage(err, t("workspace:invite.joinFailed")));
    }
  };

  const handleLoginAndJoin = async (provider: "google" | "github") => {
    if (!code) return;
    // Store the invite code in localStorage so we can auto-accept after SSO
    localStorage.setItem(PENDING_INVITE_KEY, code);
    try {
      const { url } = await getAuthURL(provider);
      window.location.href = url;
    } catch (err) {
      toast.error(getErrorMessage(err, t("auth:loginFailed")));
      localStorage.removeItem(PENDING_INVITE_KEY);
    }
  };

  if (isLoading || authLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="space-y-4 w-80">
          <Skeleton className="h-8 w-48" />
          <Skeleton className="h-32 w-full" />
        </div>
      </div>
    );
  }

  if (isError || !preview) {
    return (
      <div className="flex min-h-screen items-center justify-center p-4">
        <Card className="w-full max-w-md">
          <CardContent className="pt-6 text-center space-y-4">
            <p className="text-destructive">{t("workspace:invite.notFound")}</p>
            <a href="/login" className="text-sm underline">
              {t("workspace:invite.goToLogin")}
            </a>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (preview.expired || preview.used) {
    return (
      <div className="flex min-h-screen items-center justify-center p-4">
        <Card className="w-full max-w-md">
          <CardContent className="pt-6 text-center space-y-4">
            <p className="text-destructive">
              {preview.expired
                ? t("workspace:invite.expired")
                : t("workspace:invite.used")}
            </p>
            <a href="/login" className="text-sm underline">
              {t("workspace:invite.goToLogin")}
            </a>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <CardTitle>{t("workspace:invite.invitedTitle")}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="text-center space-y-2">
            <p className="text-lg font-semibold">{preview.workspace_name}</p>
            <p className="text-sm text-muted-foreground">
              {preview.inviter_name} {t("workspace:invite.invitedBy")}
            </p>
            <Badge variant="secondary" className="text-sm">
              {preview.role}
            </Badge>
          </div>

          {user ? (
            <Button
              className="w-full"
              onClick={handleAccept}
              disabled={acceptInvite.isPending}
            >
              {acceptInvite.isPending ? t("workspace:invite.joining") : t("workspace:invite.joinWorkspace")}
            </Button>
          ) : (
            <div className="space-y-2">
              <p className="text-sm text-center text-muted-foreground">
                {t("workspace:invite.signInToJoin")}
              </p>
              <Button
                className="w-full"
                variant="outline"
                onClick={() => handleLoginAndJoin("google")}
              >
                {t("auth:continueGoogle")}
              </Button>
              <Button
                className="w-full"
                variant="outline"
                onClick={() => handleLoginAndJoin("github")}
              >
                {t("auth:continueGithub")}
              </Button>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
