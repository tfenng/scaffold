import axios from "axios";
import type { User, Page, UserListFilter } from "@/types";

const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080",
});

type CreateUserPayload = {
  uid: string;
  email?: string;
  name: string;
  used_name?: string;
  company?: string;
  birth?: string;
};

type UpdateUserPayload = {
  email?: string;
  name: string;
  used_name?: string;
  company?: string;
  birth?: string;
};

export const userApi = {
  getById: (id: number) => api.get<User>(`/users/${id}`),

  list: (filter: UserListFilter) =>
    api.get<Page<User>>("/users", { params: filter }),

  create: (data: CreateUserPayload) => api.post<User>("/users", data),

  update: (id: number, data: UpdateUserPayload) => api.put<User>(`/users/${id}`, data),

  delete: (id: number) => api.delete<void>(`/users/${id}`),
};

export default api;
