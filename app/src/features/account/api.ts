import { apiClient } from "@/lib/api/client";
import type {
  AccountStatus,
  AccountView,
  RegisterAccountRequest,
  UpdateAccountStatusRequest,
  UpdateMyProfileRequest,
} from "@/types/account";

// POST /accounts
export function registerAccount(body: RegisterAccountRequest) {
  return apiClient.post<AccountView>("/accounts", body);
}

// GET /accounts/me
export function getMyAccount() {
  return apiClient.get<AccountView>("/accounts/me");
}

// PATCH /accounts/me
export function updateMyProfile(body: UpdateMyProfileRequest) {
  return apiClient.patch<AccountView>("/accounts/me", body);
}

// DELETE /accounts/me
export function deleteMyAccount() {
  return apiClient.delete<void>("/accounts/me");
}

// GET /accounts?status=... (admin専用)
export function listAccountsByStatus(status: AccountStatus) {
  return apiClient.get<AccountView[]>("/accounts", { searchParams: { status } });
}

// PATCH /accounts/:id/status (admin専用)
export function updateAccountStatus(accountId: string, body: UpdateAccountStatusRequest) {
  return apiClient.patch<AccountView>(`/accounts/${accountId}/status`, body);
}
