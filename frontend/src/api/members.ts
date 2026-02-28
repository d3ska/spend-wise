import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { MemberWithProfile } from "@/types";
import client from "./client";

export function useListMembers(workspaceId: number) {
  return useQuery({
    queryKey: ["workspaces", workspaceId, "members"],
    queryFn: async () => {
      const { data } = await client.get<MemberWithProfile[]>(
        `/workspaces/${workspaceId}/members`,
      );
      return data;
    },
    enabled: workspaceId > 0,
  });
}

export function useUpdateMemberRole(workspaceId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: { userId: number; role: string }) => {
      await client.put(
        `/workspaces/${workspaceId}/members/${body.userId}`,
        { role: body.role },
      );
    },
    onSuccess: () =>
      qc.invalidateQueries({
        queryKey: ["workspaces", workspaceId, "members"],
      }),
  });
}

export function useRemoveMember(workspaceId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (userId: number) => {
      await client.delete(`/workspaces/${workspaceId}/members/${userId}`);
    },
    onSuccess: () => {
      qc.invalidateQueries({
        queryKey: ["workspaces", workspaceId, "members"],
      });
      qc.invalidateQueries({ queryKey: ["workspaces"] });
    },
  });
}
