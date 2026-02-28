import {
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import type { BankAccount } from "@/types";
import client from "./client";

export function useUserBankAccounts() {
  return useQuery({
    queryKey: ["all-bank-accounts"],
    queryFn: async () => {
      const { data } = await client.get<BankAccount[]>("/bank-accounts");
      return data;
    },
  });
}

export function useLinkedBankAccounts(workspaceId: number) {
  return useQuery({
    queryKey: ["bank-accounts", workspaceId],
    queryFn: async () => {
      const { data } = await client.get<BankAccount[]>(
        `/workspaces/${workspaceId}/bank-accounts`,
      );
      return data;
    },
    enabled: workspaceId > 0,
  });
}

export function useLinkBankAccount(workspaceId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (bankAccountId: number) => {
      await client.post(`/workspaces/${workspaceId}/bank-accounts`, {
        bank_account_id: bankAccountId,
      });
    },
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ["bank-accounts", workspaceId] }),
  });
}

export function useUnlinkBankAccount(workspaceId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (bankAccountId: number) => {
      await client.delete(
        `/workspaces/${workspaceId}/bank-accounts/${bankAccountId}`,
      );
    },
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ["bank-accounts", workspaceId] }),
  });
}

export function useUpdateBankAccountCustomName() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({ bankAccountId, customName }: { bankAccountId: number; customName: string }) => {
      const { data } = await client.patch(`/bank-accounts/${bankAccountId}`, {
        custom_name: customName,
      });
      return data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["all-bank-accounts"] });
      qc.invalidateQueries({ queryKey: ["bank-accounts"] });
    },
  });
}

export function useTriggerSync() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (connectionId: number) => {
      await client.post(`/bank-connections/${connectionId}/sync`);
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["bank-connections"] });
      qc.invalidateQueries({ queryKey: ["bank-accounts"] });
      qc.invalidateQueries({ queryKey: ["transactions"] });
    },
  });
}
