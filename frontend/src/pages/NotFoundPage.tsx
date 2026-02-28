import { Link } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";

export default function NotFoundPage() {
  const { t } = useTranslation(["common"]);

  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="text-center space-y-4">
        <h1 className="text-4xl font-bold">{t("common:notFound.title")}</h1>
        <p className="text-muted-foreground">{t("common:notFound.message")}</p>
        <Button asChild>
          <Link to="/dashboard">{t("common:notFound.goToDashboard")}</Link>
        </Button>
      </div>
    </div>
  );
}
