import {
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import type { Rule } from "@/types";
import client from "./client";

export function useRules(workspaceId: number) {
  return useQuery({
    queryKey: ["rules", workspaceId],
    queryFn: async () => {
      const { data } = await client.get<Rule[]>(
        `/workspaces/${workspaceId}/rules`,
      );
      return data;
    },
    enabled: workspaceId > 0,
  });
}

export function useCreateRule(workspaceId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: {
      match_pattern: string;
      target_category_id: number;
      priority: number;
      amount_min?: string | null;
      amount_max?: string | null;
      bank_account_id?: number | null;
      counterparty_iban?: string | null;
    }) => {
      const { data } = await client.post<Rule>(
        `/workspaces/${workspaceId}/rules`,
        body,
      );
      return data;
    },
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ["rules", workspaceId] }),
  });
}

export function useUpdateRule(workspaceId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({
      id,
      ...body
    }: {
      id: number;
      match_pattern: string;
      target_category_id: number;
      priority: number;
      amount_min?: string | null;
      amount_max?: string | null;
      bank_account_id?: number | null;
      counterparty_iban?: string | null;
    }) => {
      const { data } = await client.put<Rule>(
        `/workspaces/${workspaceId}/rules/${id}`,
        body,
      );
      return data;
    },
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ["rules", workspaceId] }),
  });
}

export function useDeleteRule(workspaceId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: number) => {
      await client.delete(`/workspaces/${workspaceId}/rules/${id}`);
    },
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ["rules", workspaceId] }),
  });
}

export function useToggleRule(workspaceId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({
      id,
      enabled,
    }: {
      id: number;
      enabled: boolean;
    }) => {
      const { data } = await client.patch<Rule>(
        `/workspaces/${workspaceId}/rules/${id}/toggle`,
        { enabled },
      );
      return data;
    },
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ["rules", workspaceId] }),
  });
}
