import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { InvitePreview, WorkspaceInvite, Workspace } from "@/types";
import client from "./client";

export function usePreviewInvite(code: string) {
  return useQuery({
    queryKey: ["invites", code],
    queryFn: async () => {
      const { data } = await client.get<InvitePreview>(`/invites/${code}`);
      return data;
    },
    enabled: !!code,
  });
}

export function useAcceptInvite() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (code: string) => {
      const { data } = await client.post<Workspace>(
        `/invites/${code}/accept`,
      );
      return data;
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ["workspaces"] }),
  });
}

export function useCreateInvite(workspaceId: number) {
  return useMutation({
    mutationFn: async (body: { role: string; expires_in_hours?: number }) => {
      const { data } = await client.post<WorkspaceInvite>(
        `/workspaces/${workspaceId}/invites`,
        body,
      );
      return data;
    },
  });
}
