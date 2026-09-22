import { apiClient } from "@/lib/api/client";
import type {
  MembershipRemovalRequestView,
  MembershipView,
  RemoveMemberResponse,
} from "@/types/membership";

// POST /accounts/:id/members (owner専用)
export function addMember(accountId: string, userId: string) {
  return apiClient.post<MembershipView>(`/accounts/${accountId}/members`, { user_id: userId });
}

// GET /accounts/:id/members (owner・admin専用)
export function listMembers(accountId: string) {
  return apiClient.get<MembershipView[]>(`/accounts/${accountId}/members`);
}

// PATCH /accounts/:id/members/:user_id (owner専用、role=adminの付与のみ)
export function grantAdmin(accountId: string, userId: string) {
  return apiClient.patch<MembershipView>(`/accounts/${accountId}/members/${userId}`, {
    role: "admin",
  });
}

// DELETE /accounts/:id/members/:user_id
// ownerが実行した場合は即座に反映(200)、adminが実行した場合は承認待ち(202)。
export function removeMember(accountId: string, userId: string) {
  return apiClient.delete<RemoveMemberResponse>(`/accounts/${accountId}/members/${userId}`);
}

// GET /accounts/:id/removal-requests (owner専用)
export function listRemovalRequests(accountId: string) {
  return apiClient.get<MembershipRemovalRequestView[]>(`/accounts/${accountId}/removal-requests`);
}

// POST /accounts/:id/removal-requests/:request_id/approve (owner専用)
export function approveRemovalRequest(accountId: string, requestId: string) {
  return apiClient.post<MembershipView>(
    `/accounts/${accountId}/removal-requests/${requestId}/approve`,
  );
}

// POST /accounts/:id/removal-requests/:request_id/reject (owner専用)
export function rejectRemovalRequest(accountId: string, requestId: string) {
  return apiClient.post<MembershipRemovalRequestView>(
    `/accounts/${accountId}/removal-requests/${requestId}/reject`,
  );
}
