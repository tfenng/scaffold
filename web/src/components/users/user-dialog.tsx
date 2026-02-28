"use client";

import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { userApi } from "@/lib/api";
import { createUserSchema, updateUserSchema, type CreateUserInput, type UpdateUserInput } from "@/schemas/user-schema";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { toast } from "sonner";

interface UserDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  userId: number | null;
}

export function UserDialog({ open, onOpenChange, userId }: UserDialogProps) {
  const queryClient = useQueryClient();
  const isEdit = userId !== null;

  const { data: user, isLoading } = useQuery({
    queryKey: ["user", userId],
    queryFn: () => userApi.getById(userId!),
    enabled: isEdit && open,
  });

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CreateUserInput | UpdateUserInput>({
    resolver: zodResolver(isEdit ? updateUserSchema : createUserSchema),
  });

  useEffect(() => {
    if (user?.data && isEdit) {
      reset({
        name: user.data.name,
        used_name: user.data.used_name || "",
        company: user.data.company || "",
        birth: user.data.birth || "",
      });
    } else if (!isEdit) {
      reset({
        uid: "",
        name: "",
        used_name: "",
        company: "",
        birth: "2000-01-01",
      });
    }
  }, [user, isEdit, reset, open]);

  const createMutation = useMutation({
    mutationFn: (data: CreateUserInput) => userApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["users"] });
      toast.success("User created successfully");
      onOpenChange(false);
    },
    onError: () => {
      toast.error("Failed to create user");
    },
  });

  const updateMutation = useMutation({
    mutationFn: (data: UpdateUserInput) => userApi.update(userId!, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["users"] });
      queryClient.invalidateQueries({ queryKey: ["user", userId] });
      toast.success("User updated successfully");
      onOpenChange(false);
    },
    onError: () => {
      toast.error("Failed to update user");
    },
  });

  const onSubmit = (data: CreateUserInput | UpdateUserInput) => {
    if (isEdit) {
      updateMutation.mutate(data as UpdateUserInput);
    } else {
      createMutation.mutate(data as CreateUserInput);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{isEdit ? "Edit User" : "Create User"}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit(onSubmit)}>
          <div className="grid gap-4 py-4">
            {isEdit && (
              <div className="grid grid-cols-4 items-center gap-4">
                <Label className="text-right">UID</Label>
                <span className="col-span-3 text-muted-foreground">{user?.data?.uid || "-"}</span>
              </div>
            )}
            {!isEdit && (
              <div className="grid grid-cols-4 items-center gap-4">
                <Label htmlFor="uid" className="text-right">
                  UID
                </Label>
                <Input
                  id="uid"
                  {...register("uid")}
                  className="col-span-3"
                  placeholder="Required"
                />
                {errors.uid && (
                  <p className="col-span-4 text-right text-sm text-destructive">
                    {errors.uid.message as string}
                  </p>
                )}
              </div>
            )}
            <div className="grid grid-cols-4 items-center gap-4">
              <Label htmlFor="name" className="text-right">
                Name
              </Label>
              <Input
                id="name"
                {...register("name")}
                className="col-span-3"
                placeholder="John Doe"
              />
              {errors.name && (
                <p className="col-span-4 text-right text-sm text-destructive">
                  {errors.name.message as string}
                </p>
              )}
            </div>
            <div className="grid grid-cols-4 items-center gap-4">
              <Label htmlFor="used_name" className="text-right">
                Used Name
              </Label>
              <Input
                id="used_name"
                {...register("used_name")}
                className="col-span-3"
                placeholder="Optional"
              />
            </div>
            <div className="grid grid-cols-4 items-center gap-4">
              <Label htmlFor="company" className="text-right">
                Company
              </Label>
              <Input
                id="company"
                {...register("company")}
                className="col-span-3"
                placeholder="Optional"
              />
            </div>
            <div className="grid grid-cols-4 items-center gap-4">
              <Label htmlFor="birth" className="text-right">
                Birth
              </Label>
              <Input
                id="birth"
                type="date"
                {...register("birth")}
                className="col-span-3"
                placeholder="YYYY-MM-DD"
              />
              {errors.birth && (
                <p className="col-span-4 text-right text-sm text-destructive">
                  {errors.birth.message as string}
                </p>
              )}
            </div>
            {!isEdit && (
              <div className="grid grid-cols-4 items-center gap-4">
                <Label htmlFor="email" className="text-right">
                  Email
                </Label>
                <Input
                  id="email"
                  {...register("email")}
                  className="col-span-3"
                  placeholder="user@example.com"
                />
                {errors.email && (
                  <p className="col-span-4 text-right text-sm text-destructive">
                    {errors.email.message as string}
                  </p>
                )}
              </div>
            )}
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={createMutation.isPending || updateMutation.isPending}>
              {createMutation.isPending || updateMutation.isPending ? "Saving..." : "Save"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
