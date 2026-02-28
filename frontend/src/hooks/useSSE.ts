import { useEffect } from "react";
import { useQueryClient } from "@tanstack/react-query";

const EVENT_KEY_MAP: Record<string, (wsId: number) => (string | number)[][]> = {
  workspace_changed: () => [["workspaces"]],
  member_changed: (wsId) => [
    ["workspaces", wsId, "members"],
    ["workspaces"],
  ],
  transaction_changed: (wsId) => [
    ["transactions", wsId],
    ["summary", wsId],
  ],
  category_changed: (wsId) => [["categories", wsId]],
  funding_changed: (wsId) => [
    ["fundings", wsId],
    ["summary", wsId],
  ],
  rule_changed: (wsId) => [["rules", wsId]],
};

export function useSSE(workspaceId: number) {
  const queryClient = useQueryClient();

  useEffect(() => {
    if (!workspaceId) return;

    const url = `/api/v1/workspaces/${workspaceId}/events`;
    const source = new EventSource(url, { withCredentials: true });

    for (const eventType of Object.keys(EVENT_KEY_MAP)) {
      source.addEventListener(eventType, () => {
        const keys = EVENT_KEY_MAP[eventType](workspaceId);
        for (const key of keys) {
          queryClient.invalidateQueries({ queryKey: key });
        }
      });
    }

    return () => {
      source.close();
    };
  }, [workspaceId, queryClient]);
}
