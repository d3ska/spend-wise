import {
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import type { Funding, Money } from "@/types";
import client from "./client";

export function useFundings(workspaceId: number) {
  return useQuery({
    queryKey: ["fundings", workspaceId],
    queryFn: async () => {
      const { data } = await client.get<Funding[]>(
        `/workspaces/${workspaceId}/fundings`,
      );
      return data;
    },
    enabled: workspaceId > 0,
  });
}

export function useBudgets(workspaceId: number, month: string) {
  return useQuery({
    queryKey: ["fundings", workspaceId, month],
    queryFn: async () => {
      const { data } = await client.get<Funding[]>(
        `/workspaces/${workspaceId}/fundings`,
        { params: { from: month, to: month } },
      );
      return data;
    },
    enabled: workspaceId > 0 && month.length > 0,
  });
}

export function useUpsertBudget(workspaceId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: {
      year_month: string;
      category_id?: number | null;
      amount: Money;
    }) => {
      const { data } = await client.put<Funding>(
        `/workspaces/${workspaceId}/fundings`,
        body,
      );
      return data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["fundings", workspaceId] });
      qc.invalidateQueries({ queryKey: ["summary", workspaceId] });
    },
  });
}

export function useDeleteBudget(workspaceId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (fundingId: number) => {
      await client.delete(
        `/workspaces/${workspaceId}/fundings/${fundingId}`,
      );
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["fundings", workspaceId] });
      qc.invalidateQueries({ queryKey: ["summary", workspaceId] });
    },
  });
}

export function useRecordFunding(workspaceId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: {
      user_id: number;
      year_month: string;
      amount: Money;
    }) => {
      const { data } = await client.put<Funding>(
        `/workspaces/${workspaceId}/fundings`,
        body,
      );
      return data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["fundings", workspaceId] });
      qc.invalidateQueries({ queryKey: ["summary", workspaceId] });
    },
  });
}
