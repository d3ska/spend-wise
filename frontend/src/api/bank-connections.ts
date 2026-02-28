import {
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import type { BankAccount, BankConnection, Institution } from "@/types";
import client from "./client";

export function useInstitutions(country: string) {
  return useQuery({
    queryKey: ["institutions", country],
    queryFn: async () => {
      const { data } = await client.get<Institution[]>("/banks", {
        params: { country },
      });
      return data;
    },
    enabled: country.length > 0,
  });
}

export function useListConnections() {
  return useQuery({
    queryKey: ["bank-connections"],
    queryFn: async () => {
      const { data } = await client.get<BankConnection[]>("/bank-connections");
      return data;
    },
  });
}

export function useInitiateConnection() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: {
      institution_id: string;
      institution_name: string;
      redirect_url: string;
      country: string;
    }) => {
      const { data } = await client.post<{
        connection_id: number;
        auth_link: string;
      }>("/bank-connections", body);
      return data;
    },
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ["bank-connections"] }),
  });
}

export function useCompleteConnection() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (params: { code: string }) => {
      const { data } = await client.post<BankAccount[]>(
        "/bank-connections/complete",
        { code: params.code },
      );
      return data;
    },
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ["bank-connections"] }),
  });
}

export function useDeleteConnection() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (connectionId: number) => {
      await client.delete(`/bank-connections/${connectionId}`);
    },
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ["bank-connections"] }),
  });
}

export function useReconnectConnection() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: {
      connection_id: number;
      redirect_url: string;
    }) => {
      const { data } = await client.post<{ auth_link: string }>(
        `/bank-connections/${body.connection_id}/reconnect`,
        { redirect_url: body.redirect_url },
      );
      return data;
    },
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ["bank-connections"] }),
  });
}
