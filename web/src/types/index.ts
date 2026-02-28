export interface User {
  id: number;
  email: string;
  name: string;
  used_name: string | null;
  company: string | null;
  birth: string | null;
  created_at: string;
  updated_at: string;
}

export interface Page<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface UserListFilter {
  email?: string;
  name_like?: string;
  page?: number;
  page_size?: number;
}

export interface ApiError {
  code: string;
  message: string;
}
