import { z } from "zod";

export const createUserSchema = z.object({
  email: z.string().email("Invalid email address"),
  name: z.string().min(1, "Name is required"),
  used_name: z.string().optional(),
  company: z.string().optional(),
  birth: z.string().regex(/^\d{4}-\d{2}-\d{2}$/, "Invalid date format").optional().or(z.literal("")),
});

export const updateUserSchema = z.object({
  name: z.string().min(1, "Name is required"),
  used_name: z.string().optional(),
  company: z.string().optional(),
  birth: z.string().regex(/^\d{4}-\d{2}-\d{2}$/, "Invalid date format").optional().or(z.literal("")),
});

export const userListQuerySchema = z.object({
  email: z.string().optional(),
  name_like: z.string().optional(),
  page: z.coerce.number().min(1).default(1),
  page_size: z.coerce.number().min(1).max(200).default(20),
});

export type CreateUserInput = z.infer<typeof createUserSchema>;
export type UpdateUserInput = z.infer<typeof updateUserSchema>;
export type UserListQuery = z.infer<typeof userListQuerySchema>;
