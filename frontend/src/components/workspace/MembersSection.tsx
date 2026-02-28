import { useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { UserPlus, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog";
import { useListMembers, useUpdateMemberRole, useRemoveMember } from "@/api/members";
import { InviteDialog } from "./InviteDialog";
import { getErrorMessage } from "@/lib/errors";
import { initials } from "@/lib/format";
import type { MemberWithProfile } from "@/types";

interface MembersSectionProps {
  workspaceId: number;
  isOwner: boolean;
  currentUserId: number;
}

export function MembersSection({
  workspaceId,
  isOwner,
  currentUserId,
}: MembersSectionProps) {
  const { t } = useTranslation(["workspace", "common"]);
  const { data: members } = useListMembers(workspaceId);
  const updateRole = useUpdateMemberRole(workspaceId);
  const removeMember = useRemoveMember(workspaceId);
  const [inviteOpen, setInviteOpen] = useState(false);
  const [removeTarget, setRemoveTarget] = useState<MemberWithProfile | null>(null);

  const handleRoleChange = async (userId: number, role: string) => {
    try {
      await updateRole.mutateAsync({ userId, role });
      toast.success(t("workspace:members.roleUpdated"));
    } catch (err) {
      toast.error(getErrorMessage(err, t("workspace:members.roleUpdateFailed")));
    }
  };

  const handleRemove = async () => {
    if (!removeTarget) return;
    try {
      await removeMember.mutateAsync(removeTarget.user_id);
      toast.success(t("workspace:members.removed"));
      setRemoveTarget(null);
    } catch (err) {
      toast.error(getErrorMessage(err, t("workspace:members.removeFailed")));
    }
  };

  return (
    <>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base">{t("workspace:members.title")}</CardTitle>
          {isOwner && (
            <Button size="sm" onClick={() => setInviteOpen(true)}>
              <UserPlus className="h-4 w-4 mr-1" /> {t("workspace:members.invite")}
            </Button>
          )}
        </CardHeader>
        <CardContent className="space-y-3">
          {members && members.length > 0 ? (
            members.map((m: MemberWithProfile) => (
              <div
                key={m.user_id}
                className="flex items-center justify-between rounded-md border p-3"
              >
                <div className="flex items-center gap-3">
                  <Avatar className="h-8 w-8">
                    <AvatarImage src={m.avatar_url} />
                    <AvatarFallback>{initials(m.display_name)}</AvatarFallback>
                  </Avatar>
                  <div>
                    <p className="text-sm font-medium">
                      {m.display_name}
                      {m.user_id === currentUserId && (
                        <span className="text-muted-foreground ml-1">{t("workspace:members.you")}</span>
                      )}
                    </p>
                    <p className="text-xs text-muted-foreground">{m.email}</p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  {isOwner && m.role !== "owner" && m.user_id !== currentUserId ? (
                    <>
                      <Select
                        value={m.role}
                        onValueChange={(val) => handleRoleChange(m.user_id, val)}
                      >
                        <SelectTrigger className="w-24 h-8 text-xs">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="editor">{t("workspace:members.editor")}</SelectItem>
                          <SelectItem value="viewer">{t("workspace:members.viewer")}</SelectItem>
                        </SelectContent>
                      </Select>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => setRemoveTarget(m)}
                        title={t("common:remove")}
                      >
                        <Trash2 className="h-4 w-4 text-destructive" />
                      </Button>
                    </>
                  ) : (
                    <Badge variant="secondary" className="text-xs">
                      {m.role}
                    </Badge>
                  )}
                </div>
              </div>
            ))
          ) : (
            <p className="text-sm text-muted-foreground">{t("workspace:members.noMembers")}</p>
          )}
        </CardContent>
      </Card>

      <InviteDialog
        workspaceId={workspaceId}
        open={inviteOpen}
        onOpenChange={setInviteOpen}
      />

      <ConfirmDialog
        open={!!removeTarget}
        onOpenChange={(open) => !open && setRemoveTarget(null)}
        title={t("workspace:members.removeTitle")}
        description={`${removeTarget?.display_name} — ${t("workspace:members.removeDescription")}`}
        onConfirm={handleRemove}
        loading={removeMember.isPending}
      />
    </>
  );
}
