import { useState, useEffect, useMemo, useRef } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { Building2, Check, Link, Pencil, Plus, RefreshCw, Trash2, Unplug, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import { ConfirmDialog } from "@/components/shared/ConfirmDialog";
import { ConnectBankDialog } from "@/components/bank/ConnectBankDialog";
import { AutoLinkDialog } from "@/components/bank/AutoLinkDialog";
import { MembersSection } from "@/components/workspace/MembersSection";
import {
  useWorkspace,
  useUpdateWorkspace,
  useDeleteWorkspace,
} from "@/api/workspaces";
import {
  useListConnections,
  useDeleteConnection,
  useReconnectConnection,
} from "@/api/bank-connections";
import {
  useUserBankAccounts,
  useLinkedBankAccounts,
  useLinkBankAccount,
  useUnlinkBankAccount,
  useTriggerSync,
  useUpdateBankAccountCustomName,
} from "@/api/bank-accounts";
import { useWorkspaceContext } from "@/hooks/useWorkspace";
import { useAuth } from "@/hooks/useAuth";
import { getErrorMessage } from "@/lib/errors";
import type { BankConnection, BankAccount } from "@/types";

export default function WorkspaceSettingsPage() {
  const { t } = useTranslation(["workspace", "common"]);
  const navigate = useNavigate();
  const location = useLocation();
  const { user } = useAuth();
  const { workspaceId, workspace } = useWorkspaceContext();
  const { data: ws, isLoading } = useWorkspace(workspaceId);
  const updateMutation = useUpdateWorkspace(workspaceId);
  const deleteMutation = useDeleteWorkspace(workspaceId);

  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [initialized, setInitialized] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [connectBankOpen, setConnectBankOpen] = useState(false);
  const [autoLinkOpen, setAutoLinkOpen] = useState(false);
  const [newAccounts, setNewAccounts] = useState<BankAccount[]>([]);
  const [editingAccountId, setEditingAccountId] = useState<number | null>(null);
  const [editingName, setEditingName] = useState("");
  const editInputRef = useRef<HTMLInputElement>(null);

  const { data: connections } = useListConnections();
  const { data: allAccounts } = useUserBankAccounts();
  const { data: linkedAccounts } = useLinkedBankAccounts(workspaceId);
  const linkAccount = useLinkBankAccount(workspaceId);
  const unlinkAccount = useUnlinkBankAccount(workspaceId);
  const deleteConnection = useDeleteConnection();
  const reconnect = useReconnectConnection();
  const triggerSync = useTriggerSync();
  const renameMutation = useUpdateBankAccountCustomName();

  const isOwner = ws && user ? ws.owner_id === user.id : false;
  const canManageAccounts = isOwner;

  const accountsByConnection = useMemo(() => {
    const map = new Map<number, BankAccount[]>();
    for (const acct of allAccounts ?? []) {
      const list = map.get(acct.bank_connection_id) ?? [];
      list.push(acct);
      map.set(acct.bank_connection_id, list);
    }
    return map;
  }, [allAccounts]);

  // Open auto-link dialog when arriving from bank callback with discovered accounts.
  useEffect(() => {
    const state = location.state as { newAccounts?: BankAccount[] } | null;
    if (state?.newAccounts && state.newAccounts.length > 0) {
      setNewAccounts(state.newAccounts);
      setAutoLinkOpen(true);
      // Clear router state so refresh doesn't re-open the dialog.
      window.history.replaceState({}, document.title);
    }
  }, [location.state]);

  if (ws && !initialized) {
    setName(ws.name);
    setDescription(ws.description);
    setInitialized(true);
  }

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await updateMutation.mutateAsync({ name, description });
      toast.success(t("workspace:updated"));
    } catch (err) {
      toast.error(getErrorMessage(err, t("workspace:updateFailed")));
    }
  };

  const handleDeleteWorkspace = async () => {
    try {
      await deleteMutation.mutateAsync();
      toast.success(t("workspace:deleted"));
      localStorage.removeItem("sw_workspace_id");
      navigate("/dashboard", { replace: true });
    } catch (err) {
      toast.error(getErrorMessage(err, t("workspace:deleteFailed")));
    }
  };

  const linkedIds = new Set(
    (linkedAccounts ?? []).map((a: BankAccount) => a.id),
  );

  if (isLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-32 w-full" />
      </div>
    );
  }

  return (
    <div className="space-y-6 max-w-2xl">
      <h1 className="text-2xl font-bold">{t("workspace:title")}</h1>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">{t("workspace:details")}</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSave} className="space-y-4">
            <div className="space-y-1">
              <Label>{t("workspace:name")}</Label>
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                disabled={!isOwner}
                required
              />
            </div>
            <div className="space-y-1">
              <Label>{t("workspace:description")}</Label>
              <Input
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                disabled={!isOwner}
              />
            </div>
            {isOwner && (
              <Button type="submit" disabled={updateMutation.isPending}>
                {updateMutation.isPending ? t("common:saving") : t("common:save")}
              </Button>
            )}
          </form>
        </CardContent>
      </Card>

      <MembersSection
        workspaceId={workspaceId}
        isOwner={isOwner}
        currentUserId={user?.id ?? 0}
      />

      {/* Bank Connections with nested accounts */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base">{t("workspace:bank.title")}</CardTitle>
          <Button size="sm" onClick={() => setConnectBankOpen(true)}>
            <Plus className="h-4 w-4 mr-1" /> {t("workspace:bank.connect")}
          </Button>
        </CardHeader>
        <CardContent className="space-y-3">
          {connections && connections.length > 0 ? (
            connections.map((conn: BankConnection) => {
              const connAccounts = accountsByConnection.get(conn.id) ?? [];
              const connLinked = connAccounts.filter((a) => linkedIds.has(a.id));
              const connUnlinked = connAccounts.filter((a) => !linkedIds.has(a.id));
              return (
                <div key={conn.id}>
                  <div className="flex items-center justify-between rounded-md border p-3">
                    <div className="flex items-center gap-3">
                      <Building2 className="h-5 w-5 text-muted-foreground" />
                      <div>
                        <p className="text-sm font-medium">
                          {conn.institution_name}
                        </p>
                        <div className="flex items-center gap-2 mt-0.5">
                          <Badge
                            variant={
                              conn.expiry_status === "ok"
                                ? "default"
                                : conn.expiry_status === "warning"
                                  ? "secondary"
                                  : "destructive"
                            }
                          >
                            {conn.expiry_status === "expired"
                              ? t("workspace:bank.expired")
                              : `${conn.days_remaining}${t("workspace:bank.daysRemaining")}`}
                          </Badge>
                        </div>
                      </div>
                    </div>
                    <div className="flex items-center gap-1">
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() =>
                          triggerSync.mutate(conn.id, {
                            onSuccess: () => toast.success(t("workspace:bank.syncStarted")),
                            onError: (err) => toast.error(getErrorMessage(err, t("workspace:bank.syncFailed"))),
                          })
                        }
                        disabled={
                          conn.expiry_status === "expired" ||
                          triggerSync.isPending
                        }
                        title={t("workspace:bank.syncTransactions")}
                      >
                        <RefreshCw className="h-4 w-4" />
                      </Button>
                      {conn.expiry_status !== "ok" && (
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => {
                            const redirectUrl = `${window.location.origin}/bank/callback`;
                            reconnect.mutate(
                              {
                                connection_id: conn.id,
                                redirect_url: redirectUrl,
                              },
                              {
                                onSuccess: (data) => {
                                  window.location.href = data.auth_link;
                                },
                                onError: (err) =>
                                  toast.error(getErrorMessage(err, t("workspace:bank.reconnectFailed"))),
                              },
                            );
                          }}
                          disabled={reconnect.isPending}
                        >
                          {t("workspace:bank.reconnect")}
                        </Button>
                      )}
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() =>
                          deleteConnection.mutate(conn.id, {
                            onSuccess: () =>
                              toast.success(t("workspace:bank.connectionRemoved")),
                            onError: (err) =>
                              toast.error(getErrorMessage(err, t("workspace:bank.connectionRemoveFailed"))),
                          })
                        }
                        disabled={deleteConnection.isPending}
                        title={t("workspace:bank.removeConnection")}
                      >
                        <Trash2 className="h-4 w-4 text-destructive" />
                      </Button>
                    </div>
                  </div>

                  {/* Nested accounts for this connection */}
                  {connAccounts.length === 0 ? (
                    <p className="ml-8 border-l-2 border-muted pl-3 py-2 text-xs text-muted-foreground">
                      {t("workspace:bank.noAccounts")}
                    </p>
                  ) : (
                    <div className="ml-8 border-l-2 border-muted">
                      {connLinked.map((acct) => (
                        <div
                          key={acct.id}
                          className="flex items-center justify-between pl-3 py-2"
                        >
                          <div className="min-w-0 flex-1">
                            {editingAccountId === acct.id ? (
                              <form
                                className="flex items-center gap-1"
                                onSubmit={(e) => {
                                  e.preventDefault();
                                  renameMutation.mutate(
                                    { bankAccountId: acct.id, customName: editingName.trim() },
                                    {
                                      onSuccess: () => {
                                        toast.success(t("transaction:bankAccount.renamed"));
                                        setEditingAccountId(null);
                                      },
                                      onError: (err) =>
                                        toast.error(getErrorMessage(err, t("transaction:bankAccount.renameFailed"))),
                                    },
                                  );
                                }}
                              >
                                <Input
                                  ref={editInputRef}
                                  value={editingName}
                                  onChange={(e) => setEditingName(e.target.value)}
                                  placeholder={t("transaction:bankAccount.customNamePlaceholder")}
                                  className="h-7 text-sm"
                                  autoFocus
                                  onKeyDown={(e) => {
                                    if (e.key === "Escape") setEditingAccountId(null);
                                  }}
                                />
                                <Button type="submit" size="icon" variant="ghost" className="h-7 w-7" disabled={renameMutation.isPending}>
                                  <Check className="h-3.5 w-3.5" />
                                </Button>
                                <Button type="button" size="icon" variant="ghost" className="h-7 w-7" onClick={() => setEditingAccountId(null)}>
                                  <X className="h-3.5 w-3.5" />
                                </Button>
                              </form>
                            ) : (
                              <div className="flex items-center gap-1">
                                <p className="text-sm font-medium">{acct.display_name}</p>
                                {canManageAccounts && (
                                  <Button
                                    size="icon"
                                    variant="ghost"
                                    className="h-6 w-6"
                                    onClick={() => {
                                      setEditingAccountId(acct.id);
                                      setEditingName(acct.custom_name || "");
                                    }}
                                    title={t("transaction:bankAccount.customName")}
                                  >
                                    <Pencil className="h-3 w-3" />
                                  </Button>
                                )}
                              </div>
                            )}
                            <p className="text-xs text-muted-foreground">
                              {acct.currency ? acct.currency : ""}
                              {acct.iban ? ` · ${acct.iban}` : ""}
                              {acct.last_synced_at
                                ? ` · ${t("workspace:bank.lastSynced")} ${new Date(acct.last_synced_at).toLocaleDateString()}`
                                : ""}
                            </p>
                          </div>
                          {canManageAccounts && editingAccountId !== acct.id && (
                            <Button
                              size="sm"
                              variant="ghost"
                              onClick={() =>
                                unlinkAccount.mutate(acct.id, {
                                  onSuccess: () =>
                                    toast.success(t("workspace:bank.accountUnlinked")),
                                  onError: (err) =>
                                    toast.error(getErrorMessage(err, t("workspace:bank.accountUnlinkFailed"))),
                                })
                              }
                              disabled={unlinkAccount.isPending}
                              title={t("workspace:bank.unlinkAccount")}
                            >
                              <Unplug className="h-4 w-4" />
                            </Button>
                          )}
                        </div>
                      ))}
                      {canManageAccounts && connUnlinked.map((acct) => (
                        <div
                          key={acct.id}
                          className="flex items-center justify-between pl-3 py-2"
                        >
                          <div>
                            <p className="text-sm text-muted-foreground">{acct.name}</p>
                            <p className="text-xs text-muted-foreground">
                              {acct.currency ? acct.currency : ""}
                              {acct.iban ? ` · ${acct.iban}` : ""}
                            </p>
                          </div>
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() =>
                              linkAccount.mutate(acct.id, {
                                onSuccess: () =>
                                  toast.success(t("workspace:bank.accountLinked")),
                                onError: (err) =>
                                  toast.error(getErrorMessage(err, t("workspace:bank.accountLinkFailed"))),
                              })
                            }
                            disabled={linkAccount.isPending}
                            title={t("workspace:bank.linkToWorkspace")}
                          >
                            <Link className="h-4 w-4 mr-1" />
                            {t("workspace:bank.linkToWorkspace")}
                          </Button>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              );
            })
          ) : (
            <p className="text-sm text-muted-foreground">
              {t("workspace:bank.noBankConnections")}
            </p>
          )}
        </CardContent>
      </Card>

      <ConnectBankDialog
        open={connectBankOpen}
        onOpenChange={setConnectBankOpen}
      />

      <AutoLinkDialog
        open={autoLinkOpen}
        onOpenChange={setAutoLinkOpen}
        accounts={newAccounts}
        workspaceId={workspaceId}
        workspaceName={workspace?.name ?? "this workspace"}
      />

      {isOwner && (
        <>
          <Separator />
          <Card className="border-destructive">
            <CardHeader>
              <CardTitle className="text-base text-destructive">
                {t("workspace:dangerZone.title")}
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-muted-foreground mb-4">
                {t("workspace:dangerZone.deleteDescription")}
              </p>
              <Button variant="destructive" onClick={() => setDeleteOpen(true)}>
                {t("workspace:dangerZone.deleteButton")}
              </Button>
            </CardContent>
          </Card>
        </>
      )}

      <ConfirmDialog
        open={deleteOpen}
        onOpenChange={setDeleteOpen}
        title={t("workspace:dangerZone.deleteConfirmTitle")}
        description={t("workspace:dangerZone.deleteConfirmDescription")}
        onConfirm={handleDeleteWorkspace}
        loading={deleteMutation.isPending}
      />
    </div>
  );
}
