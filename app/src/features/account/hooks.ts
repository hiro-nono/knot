import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import {
  deleteMyAccount,
  getMyAccount,
  listAccountsByStatus,
  registerAccount,
  updateAccountStatus,
  updateMyProfile,
} from "@/features/account/api";
import { accountKeys } from "@/features/account/query-keys";
import type { AccountStatus, AccountStatusAction } from "@/types/account";

// GET /accounts/me
export function useMyAccount(enabled = true) {
  return useQuery({
    queryKey: accountKeys.me(),
    queryFn: getMyAccount,
    enabled,
  });
}

// POST /accounts
export function useRegisterAccount() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: registerAccount,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: accountKeys.me() });
    },
  });
}

// PATCH /accounts/me
export function useUpdateMyProfile() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: updateMyProfile,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: accountKeys.me() });
    },
  });
}

// DELETE /accounts/me
export function useDeleteMyAccount() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: deleteMyAccount,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: accountKeys.all });
    },
  });
}

// GET /accounts?status=... (admin専用)
export function useAccountsByStatus(status: AccountStatus, enabled = true) {
  return useQuery({
    queryKey: accountKeys.byStatus(status),
    queryFn: () => listAccountsByStatus(status),
    enabled,
  });
}

// PATCH /accounts/:id/status (admin専用)
export function useUpdateAccountStatus() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ accountId, action }: { accountId: string; action: AccountStatusAction }) =>
      updateAccountStatus(accountId, { action }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: accountKeys.all });
    },
  });
}
