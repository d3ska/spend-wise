import { useState } from "react";
import { Outlet } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { Header } from "./Header";
import { Sidebar } from "./Sidebar";
import {
  Sheet,
  SheetContent,
  SheetTitle,
} from "@/components/ui/sheet";
import { WorkspaceProvider, useWorkspaceContext } from "@/hooks/useWorkspace";
import { DateRangeProvider } from "@/hooks/useDateRange";
import { CreateWorkspacePrompt } from "./CreateWorkspacePrompt";
import { useSSE } from "@/hooks/useSSE";

function AppShellInner() {
  const { t } = useTranslation(["common"]);
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const { needsOnboarding, workspaceId } = useWorkspaceContext();

  useSSE(workspaceId);

  if (needsOnboarding) {
    return <CreateWorkspacePrompt />;
  }

  return (
    <div className="flex h-screen flex-col">
      <Header onMenuClick={() => setSidebarOpen(true)} />
      <div className="flex flex-1 overflow-hidden">
        <aside className="hidden w-56 shrink-0 border-r md:block overflow-y-auto">
          <Sidebar />
        </aside>
        <Sheet open={sidebarOpen} onOpenChange={setSidebarOpen}>
          <SheetContent side="left" className="w-56 p-0">
            <SheetTitle className="sr-only">{t("common:navigation")}</SheetTitle>
            <Sidebar onNavigate={() => setSidebarOpen(false)} />
          </SheetContent>
        </Sheet>
        <main className="flex-1 overflow-y-auto p-4 md:p-8">
          <div className="mx-auto w-full max-w-6xl">
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  );
}

export function AppShell() {
  return (
    <WorkspaceProvider>
      <DateRangeProvider>
        <AppShellInner />
      </DateRangeProvider>
    </WorkspaceProvider>
  );
}
