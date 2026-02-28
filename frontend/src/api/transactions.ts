import {
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import type { Money, Transaction, TransactionType } from "@/types";
import client from "./client";

export interface TransactionListParams {
  from?: string;
  to?: string;
  limit?: number;
  offset?: number;
  type?: TransactionType;
  currency?: string;
  category_id?: number;
  bank_account_id?: number;
  amount_min?: string;
  amount_max?: string;
  sort_by?: "date" | "amount" | "description";
  sort_order?: "asc" | "desc";
}

export function useTransactions(
  workspaceId: number,
  params: TransactionListParams = {},
) {
  return useQuery({
    queryKey: ["transactions", workspaceId, params],
    queryFn: async () => {
      const { data } = await client.get<Transaction[]>(
        `/workspaces/${workspaceId}/transactions`,
        { params },
      );
      return data;
    },
    enabled: workspaceId > 0,
  });
}

export function useTransaction(workspaceId: number, txId: number) {
  return useQuery({
    queryKey: ["transactions", workspaceId, txId],
    queryFn: async () => {
      const { data } = await client.get<Transaction>(
        `/workspaces/${workspaceId}/transactions/${txId}`,
      );
      return data;
    },
    enabled: workspaceId > 0 && txId > 0,
  });
}

export interface CreateTransactionBody {
  description: string;
  date: string;
  total_amount: Money;
  type?: TransactionType;
  notes?: string;
  entries: {
    category_id?: number;
    participant_id?: number | null;
    amount: Money;
    note?: string;
  }[];
}

export function useCreateTransaction(workspaceId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: CreateTransactionBody) => {
      const { data } = await client.post<Transaction>(
        `/workspaces/${workspaceId}/transactions`,
        body,
      );
      return data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["transactions", workspaceId] });
      qc.invalidateQueries({ queryKey: ["summary", workspaceId] });
    },
  });
}

export function useDeleteTransaction(workspaceId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (txId: number) => {
      await client.delete(`/workspaces/${workspaceId}/transactions/${txId}`);
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["transactions", workspaceId] });
      qc.invalidateQueries({ queryKey: ["summary", workspaceId] });
    },
  });
}

export function useUpdateTransaction(workspaceId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({
      txId,
      body,
    }: {
      txId: number;
      body: CreateTransactionBody;
    }) => {
      const { data } = await client.put<Transaction>(
        `/workspaces/${workspaceId}/transactions/${txId}`,
        body,
      );
      return data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["transactions", workspaceId] });
      qc.invalidateQueries({ queryKey: ["summary", workspaceId] });
    },
  });
}

export function useApplyRules(workspaceId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async () => {
      const { data } = await client.post<{ updated: number; total: number }>(
        `/workspaces/${workspaceId}/transactions/apply-rules`,
      );
      return data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["transactions", workspaceId] });
      qc.invalidateQueries({ queryKey: ["summary", workspaceId] });
    },
  });
}

export function useBulkDeleteTransactions(workspaceId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (ids: number[]) => {
      const { data } = await client.post<{ deleted: number }>(
        `/workspaces/${workspaceId}/transactions/bulk-delete`,
        { ids },
      );
      return data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["transactions", workspaceId] });
      qc.invalidateQueries({ queryKey: ["summary", workspaceId] });
    },
  });
}

export function useBulkCategorizeTransactions(workspaceId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: { ids: number[]; category_id: number }) => {
      const { data } = await client.post<{ updated: number }>(
        `/workspaces/${workspaceId}/transactions/bulk-categorize`,
        body,
      );
      return data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["transactions", workspaceId] });
      qc.invalidateQueries({ queryKey: ["summary", workspaceId] });
    },
  });
}
