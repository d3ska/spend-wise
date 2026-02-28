import { useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { Copy, Check } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useCreateInvite } from "@/api/invites";
import { getErrorMessage } from "@/lib/errors";

interface InviteDialogProps {
  workspaceId: number;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function InviteDialog({
  workspaceId,
  open,
  onOpenChange,
}: InviteDialogProps) {
  const { t } = useTranslation(["workspace"]);
  const [role, setRole] = useState("viewer");
  const [inviteUrl, setInviteUrl] = useState("");
  const [copied, setCopied] = useState(false);
  const createInvite = useCreateInvite(workspaceId);

  const handleGenerate = async () => {
    try {
      const invite = await createInvite.mutateAsync({ role });
      const url = `${window.location.origin}/invite/${invite.code}`;
      setInviteUrl(url);
      setCopied(false);
    } catch (err) {
      toast.error(getErrorMessage(err, t("workspace:invite.generateFailed")));
    }
  };

  const handleCopy = async () => {
    await navigator.clipboard.writeText(inviteUrl);
    setCopied(true);
    toast.success(t("workspace:invite.linkCopied"));
    setTimeout(() => setCopied(false), 2000);
  };

  const handleClose = (isOpen: boolean) => {
    if (!isOpen) {
      setInviteUrl("");
      setCopied(false);
      setRole("viewer");
    }
    onOpenChange(isOpen);
  };

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t("workspace:invite.title")}</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          <div className="space-y-2">
            <Label>{t("workspace:invite.role")}</Label>
            <Select value={role} onValueChange={setRole}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="viewer">{t("workspace:invite.viewerDescription")}</SelectItem>
                <SelectItem value="editor">{t("workspace:invite.editorDescription")}</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {!inviteUrl ? (
            <Button
              className="w-full"
              onClick={handleGenerate}
              disabled={createInvite.isPending}
            >
              {createInvite.isPending ? t("workspace:invite.generating") : t("workspace:invite.generateLink")}
            </Button>
          ) : (
            <div className="space-y-2">
              <Label>{t("workspace:invite.linkTitle")}</Label>
              <div className="flex gap-2">
                <Input value={inviteUrl} readOnly className="text-xs" />
                <Button size="icon" variant="outline" onClick={handleCopy}>
                  {copied ? (
                    <Check className="h-4 w-4" />
                  ) : (
                    <Copy className="h-4 w-4" />
                  )}
                </Button>
              </div>
              <p className="text-xs text-muted-foreground">
                {t("workspace:invite.linkExpiry")}
              </p>
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
