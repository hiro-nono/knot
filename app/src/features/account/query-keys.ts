import type { AccountStatus } from "@/types/account";

export const accountKeys = {
  all: ["accounts"] as const,
  me: () => [...accountKeys.all, "me"] as const,
  byStatus: (status: AccountStatus) => [...accountKeys.all, "by-status", status] as const,
};
