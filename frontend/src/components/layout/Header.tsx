import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Menu, LogOut, ChevronDown, Moon, Sun, Plus, Users } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { useAuth } from "@/hooks/useAuth";
import { useWorkspaceContext } from "@/hooks/useWorkspace";
import { useTheme } from "@/hooks/useTheme";
import { CreateWorkspaceDialog } from "@/components/workspace/CreateWorkspaceDialog";
import { LanguageSwitcher } from "@/components/shared/LanguageSwitcher";
import { initials } from "@/lib/format";

export function Header({ onMenuClick }: { onMenuClick: () => void }) {
  const { t } = useTranslation(["common"]);
  const { user, logout } = useAuth();
  const { workspace, workspaces, setWorkspaceId } = useWorkspaceContext();
  const { theme, toggleTheme } = useTheme();
  const [createOpen, setCreateOpen] = useState(false);

  const ini = initials(user?.display_name ?? "");

  return (
    <header className="flex h-14 items-center gap-4 border-b bg-background px-4">
      <Button
        variant="ghost"
        size="icon"
        className="md:hidden"
        onClick={onMenuClick}
      >
        <Menu className="h-5 w-5" />
      </Button>

      <span className="text-lg font-bold">{t("common:appName")}</span>

      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="outline" size="sm" className="ml-2">
            {workspace?.name ?? t("common:selectWorkspace")}
            {workspace?.member_count ? (
              <span className="ml-1.5 flex items-center gap-0.5 text-muted-foreground">
                <Users className="h-3 w-3" />
                {workspace.member_count}
              </span>
            ) : null}
            <ChevronDown className="ml-1 h-3 w-3" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent>
          {workspaces.map((ws) => (
            <DropdownMenuItem
              key={ws.id}
              onClick={() => setWorkspaceId(ws.id)}
            >
              <span className="flex-1">{ws.name}</span>
              {ws.member_count > 0 && (
                <span className="ml-2 flex items-center gap-0.5 text-xs text-muted-foreground">
                  <Users className="h-3 w-3" />
                  {ws.member_count}
                </span>
              )}
            </DropdownMenuItem>
          ))}
          <DropdownMenuSeparator />
          <DropdownMenuItem onClick={() => setCreateOpen(true)}>
            <Plus className="mr-2 h-4 w-4" />
            {t("common:createWorkspace")}
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <div className="ml-auto flex items-center gap-1">
        <Button
          variant="ghost"
          size="icon"
          onClick={toggleTheme}
          aria-label={t("common:toggleTheme")}
        >
          {theme === "dark" ? (
            <Sun className="h-5 w-5" />
          ) : (
            <Moon className="h-5 w-5" />
          )}
        </Button>
        <LanguageSwitcher />
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon" className="rounded-full">
              <Avatar className="h-8 w-8">
                <AvatarImage src={user?.avatar_url} />
                <AvatarFallback>{ini}</AvatarFallback>
              </Avatar>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuLabel>
              <div className="flex flex-col">
                <span className="text-sm font-medium">
                  {user?.display_name}
                </span>
                <span className="text-xs text-muted-foreground">
                  {user?.email}
                </span>
              </div>
            </DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={logout}>
              <LogOut className="mr-2 h-4 w-4" />
              {t("common:logOut")}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
      <CreateWorkspaceDialog open={createOpen} onOpenChange={setCreateOpen} />
    </header>
  );
}
