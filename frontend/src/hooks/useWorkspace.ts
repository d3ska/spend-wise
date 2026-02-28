import {
  createContext,
  useContext,
  useMemo,
  type ReactNode,
} from "react";
import { createElement, useState } from "react";
import { useWorkspaces } from "@/api/workspaces";
import type { Workspace } from "@/types";

interface WorkspaceContextValue {
  workspaceId: number;
  workspace: Workspace | null;
  workspaces: Workspace[];
  isLoading: boolean;
  setWorkspaceId: (id: number) => void;
  needsOnboarding: boolean;
}

const WorkspaceContext = createContext<WorkspaceContextValue>({
  workspaceId: 0,
  workspace: null,
  workspaces: [],
  isLoading: true,
  setWorkspaceId: () => {},
  needsOnboarding: false,
});

const STORAGE_KEY = "sw_workspace_id";

export function WorkspaceProvider({ children }: { children: ReactNode }) {
  const { data: workspaces = [], isLoading } = useWorkspaces();
  const [selectedId, setSelectedId] = useState<number>(() => {
    const stored = localStorage.getItem(STORAGE_KEY);
    return stored ? Number(stored) : 0;
  });

  const setWorkspaceId = (id: number) => {
    setSelectedId(id);
    localStorage.setItem(STORAGE_KEY, String(id));
  };

  const resolved = useMemo(() => {
    if (isLoading || workspaces.length === 0) {
      return { workspaceId: selectedId, workspace: null };
    }
    const current = workspaces.find((w) => w.id === selectedId);
    if (current) {
      return { workspaceId: current.id, workspace: current };
    }
    const fallback = workspaces[0];
    // Persist the fallback selection outside of render
    queueMicrotask(() => {
      localStorage.setItem(STORAGE_KEY, String(fallback.id));
    });
    return { workspaceId: fallback.id, workspace: fallback };
  }, [workspaces, isLoading, selectedId]);

  return createElement(
    WorkspaceContext.Provider,
    {
      value: {
        workspaceId: resolved.workspaceId,
        workspace: resolved.workspace,
        workspaces,
        isLoading,
        setWorkspaceId,
        needsOnboarding: !isLoading && workspaces.length === 0,
      },
    },
    children,
  );
}

export function useWorkspaceContext() {
  return useContext(WorkspaceContext);
}
