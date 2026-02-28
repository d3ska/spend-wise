import { useQuery } from "@tanstack/react-query";
import type { Summary } from "@/types";
import client from "./client";

export function useSummary(workspaceId: number, from: string, to: string) {
  return useQuery({
    queryKey: ["summary", workspaceId, from, to],
    queryFn: async () => {
      const { data } = await client.get<Summary>(
        `/workspaces/${workspaceId}/summary`,
        { params: { from, to } },
      );
      return data;
    },
    enabled: workspaceId > 0 && !!from && !!to,
  });
}
