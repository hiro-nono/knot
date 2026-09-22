import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import {
  addMember,
  approveRemovalRequest,
  grantAdmin,
  listMembers,
  listRemovalRequests,
  rejectRemovalRequest,
  removeMember,
} from "@/features/membership/api";
import { membershipKeys } from "@/features/membership/query-keys";

// GET /accounts/:id/members
export function useMembers(accountId: string, enabled = true) {
  return useQuery({
    queryKey: membershipKeys.members(accountId),
    queryFn: () => listMembers(accountId),
    enabled,
  });
}

// POST /accounts/:id/members
export function useAddMember(accountId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (userId: string) => addMember(accountId, userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: membershipKeys.members(accountId) });
    },
  });
}

// PATCH /accounts/:id/members/:user_id
export function useGrantAdmin(accountId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (userId: string) => grantAdmin(accountId, userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: membershipKeys.members(accountId) });
    },
  });
}

// DELETE /accounts/:id/members/:user_id
export function useRemoveMember(accountId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (userId: string) => removeMember(accountId, userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: membershipKeys.members(accountId) });
      queryClient.invalidateQueries({ queryKey: membershipKeys.removalRequests(accountId) });
    },
  });
}

// GET /accounts/:id/removal-requests
export function useRemovalRequests(accountId: string, enabled = true) {
  return useQuery({
    queryKey: membershipKeys.removalRequests(accountId),
    queryFn: () => listRemovalRequests(accountId),
    enabled,
  });
}

// POST /accounts/:id/removal-requests/:request_id/approve
export function useApproveRemovalRequest(accountId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (requestId: string) => approveRemovalRequest(accountId, requestId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: membershipKeys.members(accountId) });
      queryClient.invalidateQueries({ queryKey: membershipKeys.removalRequests(accountId) });
    },
  });
}

// POST /accounts/:id/removal-requests/:request_id/reject
export function useRejectRemovalRequest(accountId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (requestId: string) => rejectRemovalRequest(accountId, requestId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: membershipKeys.removalRequests(accountId) });
    },
  });
}
