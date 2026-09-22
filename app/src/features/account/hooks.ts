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
//
// 登録直後にDashboardLayoutがGET /accounts/meを未確定のまま評価してしまう
// (invalidateQueriesは、その時点でまだ観測者が存在しないため実際の再取得を
// 保証しない)競合を避けるため、レスポンスの内容をそのままキャッシュへ
// 書き込み、DashboardLayout側の初回読み取りが必ず成功状態から始まるようにする。
export function useRegisterAccount() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: registerAccount,
    onSuccess: (account) => {
      queryClient.setQueryData(accountKeys.me(), account);
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
