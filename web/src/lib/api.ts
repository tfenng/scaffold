import axios from "axios";
import type { User, Page, UserListFilter } from "@/types";

const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080",
});

export const userApi = {
  getById: (id: number) => api.get(`/users/${id}`),

  list: (filter: UserListFilter) =>
    api.get("/users", { params: filter }),

  create: (data: {
    email: string;
    name: string;
    used_name?: string;
    company?: string;
    birth?: string;
  }) => api.post("/users", data),

  update: (id: number, data: {
    name: string;
    used_name?: string;
    company?: string;
    birth?: string;
  }) => api.put(`/users/${id}`, data),

  delete: (id: number) => api.delete(`/users/${id}`),
};

export default api;
