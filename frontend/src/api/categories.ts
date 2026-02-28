import {
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import type { Category } from "@/types";
import client from "./client";

export function useCategories(workspaceId: number) {
  return useQuery({
    queryKey: ["categories", workspaceId],
    queryFn: async () => {
      const { data } = await client.get<Category[]>(
        `/workspaces/${workspaceId}/categories`,
      );
      return data;
    },
    enabled: workspaceId > 0,
  });
}

export function useCreateCategory(workspaceId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: { name: string; icon: string }) => {
      const { data } = await client.post<Category>(
        `/workspaces/${workspaceId}/categories`,
        body,
      );
      return data;
    },
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ["categories", workspaceId] }),
  });
}

export function useUpdateCategory(workspaceId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({
      id,
      ...body
    }: {
      id: number;
      name: string;
      icon: string;
    }) => {
      const { data } = await client.put<Category>(
        `/workspaces/${workspaceId}/categories/${id}`,
        body,
      );
      return data;
    },
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ["categories", workspaceId] }),
  });
}

export function useDeleteCategory(workspaceId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: number) => {
      await client.delete(`/workspaces/${workspaceId}/categories/${id}`);
    },
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ["categories", workspaceId] }),
  });
}
