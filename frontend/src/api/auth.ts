import type { User } from "@/types";
import client from "./client";

export async function getAuthURL(provider: string) {
  const { data } = await client.get<{ url: string; state: string }>(
    `/auth/${provider}/url`,
  );
  return data;
}

export async function handleCallback(
  provider: string,
  code: string,
  state: string,
) {
  const { data } = await client.post<{ user: User }>(
    `/auth/${provider}/callback`,
    { code, state },
  );
  return data;
}

export async function getMe() {
  const { data } = await client.get<User>("/auth/me");
  return data;
}

export async function logout() {
  await client.post("/auth/logout");
}
