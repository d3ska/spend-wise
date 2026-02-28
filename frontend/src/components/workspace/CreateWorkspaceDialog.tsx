import { useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useCreateWorkspace } from "@/api/workspaces";
import { useWorkspaceContext } from "@/hooks/useWorkspace";
import { getErrorMessage } from "@/lib/errors";

interface CreateWorkspaceDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function CreateWorkspaceDialog({
  open,
  onOpenChange,
}: CreateWorkspaceDialogProps) {
  const { t } = useTranslation(["workspace", "common"]);
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const createWorkspace = useCreateWorkspace();
  const { setWorkspaceId } = useWorkspaceContext();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const ws = await createWorkspace.mutateAsync({ name, description });
      setWorkspaceId(ws.id);
      toast.success(t("workspace:created"));
      setName("");
      setDescription("");
      onOpenChange(false);
    } catch (err) {
      toast.error(getErrorMessage(err, t("workspace:createFailed")));
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t("workspace:createTitle")}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label>{t("workspace:name")}</Label>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder={t("workspace:namePlaceholder")}
              required
            />
          </div>
          <div className="space-y-2">
            <Label>{t("workspace:description")}</Label>
            <Input
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder={t("workspace:descriptionPlaceholder")}
              required
            />
          </div>
          <Button
            type="submit"
            className="w-full"
            disabled={createWorkspace.isPending}
          >
            {createWorkspace.isPending ? t("common:creating") : t("common:create")}
          </Button>
        </form>
      </DialogContent>
    </Dialog>
  );
}
