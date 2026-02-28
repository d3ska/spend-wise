import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useCreateWorkspace } from "@/api/workspaces";
import { useWorkspaceContext } from "@/hooks/useWorkspace";

export function CreateWorkspacePrompt() {
  const { t } = useTranslation(["workspace", "common"]);
  const [name, setName] = useState("");
  const create = useCreateWorkspace();
  const { setWorkspaceId } = useWorkspaceContext();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const ws = await create.mutateAsync({
      name,
      description: name,
    });
    setWorkspaceId(ws.id);
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-muted p-4">
      <Card className="w-full max-w-sm">
        <CardHeader className="text-center">
          <CardTitle>{t("workspace:welcome")}</CardTitle>
          <CardDescription>
            {t("workspace:welcomeDescription")}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="name">{t("workspace:name")}</Label>
              <Input
                id="name"
                placeholder={t("workspace:workspaceNamePlaceholder")}
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
              />
            </div>
            <Button
              type="submit"
              className="w-full"
              disabled={create.isPending || !name.trim()}
            >
              {create.isPending ? t("common:creating") : t("common:createWorkspace")}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
