import {
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import type { Workspace } from "@/types";
import client from "./client";

export function useWorkspaces() {
  return useQuery({
    queryKey: ["workspaces"],
    queryFn: async () => {
      const { data } = await client.get<Workspace[]>("/workspaces");
      return data;
    },
  });
}

export function useWorkspace(id: number) {
  return useQuery({
    queryKey: ["workspaces", id],
    queryFn: async () => {
      const { data } = await client.get<Workspace>(
        `/workspaces/${id}`,
      );
      return data;
    },
    enabled: id > 0,
  });
}

export function useCreateWorkspace() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: {
      name: string;
      description: string;
    }) => {
      const { data } = await client.post<Workspace>("/workspaces", body);
      return data;
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ["workspaces"] }),
  });
}

export function useUpdateWorkspace(id: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (body: { name: string; description: string }) => {
      const { data } = await client.put<Workspace>(
        `/workspaces/${id}`,
        body,
      );
      return data;
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ["workspaces"] }),
  });
}

export function useDeleteWorkspace(id: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async () => {
      await client.delete(`/workspaces/${id}`);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ["workspaces"] }),
  });
}
