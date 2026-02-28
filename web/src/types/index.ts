export interface User {
  ID: number;
  Email: string;
  Name: string;
  UsedName: string | null;
  Company: string | null;
  Birth: string | null;
  CreatedAt: string;
  UpdatedAt: string;
}

export interface Page<T> {
  Items: T[];
  Total: number;
  Page: number;
  PageSize: number;
  TotalPages: number;
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
