import { create } from "zustand";
import type { User } from "@/types";

interface UserState {
  selectedUser: User | null;
  setSelectedUser: (user: User | null) => void;
}

export const useUserStore = create<UserState>((set) => ({
  selectedUser: null,
  setSelectedUser: (user) => set({ selectedUser: user }),
}));
