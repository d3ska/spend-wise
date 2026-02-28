import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { getAuthURL } from "@/api/auth";

export default function LoginPage() {
  const { t } = useTranslation(["auth", "common"]);
  const [loading, setLoading] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const handleLogin = async (provider: string) => {
    setLoading(provider);
    setError(null);
    try {
      const { url } = await getAuthURL(provider);
      window.location.href = url;
    } catch {
      setError(t("auth:loginFailed"));
      setLoading(null);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-muted p-4">
      <Card className="w-full max-w-sm">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl font-bold">{t("common:appName")}</CardTitle>
          <CardDescription>
            {t("auth:signInTitle")}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          {error && (
            <p className="text-sm text-destructive text-center">{error}</p>
          )}
          <Button
            className="w-full"
            variant="outline"
            disabled={loading !== null}
            onClick={() => handleLogin("google")}
          >
            {loading === "google" ? t("auth:redirecting") : t("auth:signInGoogle")}
          </Button>
          <Button
            className="w-full"
            variant="outline"
            disabled={loading !== null}
            onClick={() => handleLogin("github")}
          >
            {loading === "github" ? t("auth:redirecting") : t("auth:signInGithub")}
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
