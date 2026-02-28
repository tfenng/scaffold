export interface User {
  ID: number;
  Uid: string;
  Email: string | null;
  Name: string;
  UsedName: string | null;
  Company: string | null;
  Birth: string | null;
  CreatedAt: string;
  UpdatedAt: string;
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
