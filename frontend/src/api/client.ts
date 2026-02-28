import axios from "axios";

const client = axios.create({
  baseURL: "/api/v1",
  withCredentials: true,
  headers: { "Content-Type": "application/json" },
});

client.interceptors.response.use(
  (response) => response,
  (error) => {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      const currentPath = window.location.pathname;
      if (currentPath !== "/login" && !currentPath.startsWith("/auth/") && !currentPath.startsWith("/invite/")) {
        window.location.href = "/login";
      }
    }
    return Promise.reject(error);
  },
);

export default client;
