"use client";

import { type FormEvent, useState } from "react";

import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Chip, type ChipTone } from "@/components/ui/chip";
import { FormAlert } from "@/components/ui/form-alert";
import { Spinner } from "@/components/ui/spinner";
import { TextField } from "@/components/ui/text-field";
import { useMyAccount } from "@/features/account/hooks";
import {
  useAddMember,
  useApproveRemovalRequest,
  useGrantAdmin,
  useMembers,
  useRejectRemovalRequest,
  useRemovalRequests,
  useRemoveMember,
} from "@/features/membership/hooks";
import { UserName } from "@/features/user/components/user-name";
import { ApiError } from "@/lib/api/errors";
import type { MembershipRole } from "@/types/membership";

const ROLE_TONE: Record<MembershipRole, ChipTone> = {
  owner: "primary",
  admin: "success",
  member: "neutral",
};

const ROLE_LABEL: Record<MembershipRole, string> = {
  owner: "オーナー",
  admin: "管理者",
  member: "メンバー",
};

// Account配下のMembership(メンバー)を管理する画面。
// GET /accounts/:id/members はowner・admin専用のため、memberが呼び出すと403になる。
// その場合はrole=memberとみなし、管理操作を行えないようにする。
export default function AccountMembersPage() {
  const myAccount = useMyAccount();
  // AccountView.user_idは、このAccountのowner MembershipにひもづくUserのID
  // (Supabaseのセッションuser idとは別の、ドメイン内部のUser ID)。
  const myUserId = myAccount.data?.user_id;
  const accountId = myAccount.data?.id;
  const hasAccountId = Boolean(accountId);

  const members = useMembers(accountId ?? "", hasAccountId);

  const membersForbidden = members.error instanceof ApiError && members.error.isForbidden;
  const myMembership = members.data?.find((member) => member.user_id === myUserId);
  const myRole: MembershipRole | undefined = membersForbidden ? "member" : myMembership?.role;

  const canAddMember = myRole === "owner";
  const canGrantAdmin = myRole === "owner";
  const canRemoveMember = myRole === "owner" || myRole === "admin";
  const canManageRemovalRequests = myRole === "owner";

  const removalRequests = useRemovalRequests(accountId ?? "", hasAccountId && canManageRemovalRequests);

  const addMember = useAddMember(accountId ?? "");
  const grantAdmin = useGrantAdmin(accountId ?? "");
  const removeMember = useRemoveMember(accountId ?? "");
  const approveRemovalRequest = useApproveRemovalRequest(accountId ?? "");
  const rejectRemovalRequest = useRejectRemovalRequest(accountId ?? "");

  const [newMemberUserId, setNewMemberUserId] = useState("");

  function handleAddMember(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!newMemberUserId.trim()) {
      return;
    }
    addMember.mutate(newMemberUserId, { onSuccess: () => setNewMemberUserId("") });
  }

  if (myAccount.isLoading) {
    return (
      <div className="flex flex-1 items-center justify-center">
        <Spinner />
      </div>
    );
  }

  if (!accountId) {
    return (
      <div className="mx-auto w-full max-w-2xl flex-1 px-4 py-10">
        <FormAlert variant="error">Accountの取得に失敗しました</FormAlert>
      </div>
    );
  }

  return (
    <div className="mx-auto flex w-full max-w-2xl flex-1 flex-col gap-6 px-4 py-10">
      <h1 className="text-xl font-semibold text-zinc-900">メンバー管理</h1>

      {membersForbidden ? (
        <FormAlert variant="error">
          メンバー一覧を閲覧する権限がありません(オーナー・管理者のみ閲覧できます)
        </FormAlert>
      ) : (
        <Card className="flex flex-col gap-3 p-4">
          {members.data?.length ? (
            <ul className="flex flex-col divide-y divide-zinc-100">
              {members.data.map((member) => (
                <li
                  key={member.id}
                  className="flex flex-wrap items-center justify-between gap-2 py-3 first:pt-0 last:pb-0"
                >
                  <div className="flex items-center gap-2">
                    <span className="text-sm">
                      <UserName userId={member.user_id} />
                    </span>
                    <Chip tone={ROLE_TONE[member.role]}>{ROLE_LABEL[member.role]}</Chip>
                    {member.status === "removed" ? <Chip tone="neutral">除外済み</Chip> : null}
                  </div>
                  <div className="flex gap-2">
                    {canGrantAdmin && member.role === "member" ? (
                      <button
                        type="button"
                        disabled={grantAdmin.isPending}
                        onClick={() => grantAdmin.mutate(member.user_id)}
                        className="text-sm font-medium text-primary hover:underline disabled:opacity-50"
                      >
                        管理者にする
                      </button>
                    ) : null}
                    {canRemoveMember && member.user_id !== myUserId && member.status === "active" ? (
                      <button
                        type="button"
                        disabled={removeMember.isPending}
                        onClick={() => removeMember.mutate(member.user_id)}
                        className="text-sm font-medium text-red-600 hover:underline disabled:opacity-50"
                      >
                        削除
                      </button>
                    ) : null}
                  </div>
                </li>
              ))}
            </ul>
          ) : (
            <p className="py-4 text-center text-sm text-zinc-500">メンバーがいません</p>
          )}
        </Card>
      )}

      {canAddMember ? (
        <form onSubmit={handleAddMember} className="flex items-end gap-2">
          <div className="flex-1">
            <TextField
              label="ユーザーIDを追加"
              placeholder="追加するuser_id"
              value={newMemberUserId}
              onChange={(event) => setNewMemberUserId(event.target.value)}
            />
          </div>
          <Button type="submit" isLoading={addMember.isPending} className="w-auto px-6">
            追加
          </Button>
        </form>
      ) : null}

      {canManageRemovalRequests && removalRequests.data?.length ? (
        <div className="flex flex-col gap-2">
          <h2 className="text-sm font-medium text-zinc-500">削除申請</h2>
          <Card className="flex flex-col gap-3 p-4">
            <ul className="flex flex-col divide-y divide-zinc-100">
              {removalRequests.data.map((request) => (
                <li
                  key={request.id}
                  className="flex flex-wrap items-center justify-between gap-2 py-3 first:pt-0 last:pb-0"
                >
                  <span className="text-sm">
                    {(() => {
                      const targetUserId = members.data?.find(
                        (m) => m.id === request.membership_id,
                      )?.user_id;
                      return targetUserId ? (
                        <UserName userId={targetUserId} />
                      ) : (
                        <span className="text-zinc-900">{request.membership_id}</span>
                      );
                    })()}
                  </span>
                  {request.status === "pending" ? (
                    <div className="flex gap-3">
                      <button
                        type="button"
                        disabled={approveRemovalRequest.isPending}
                        onClick={() => approveRemovalRequest.mutate(request.id)}
                        className="text-sm font-medium text-primary hover:underline disabled:opacity-50"
                      >
                        承認
                      </button>
                      <button
                        type="button"
                        disabled={rejectRemovalRequest.isPending}
                        onClick={() => rejectRemovalRequest.mutate(request.id)}
                        className="text-sm font-medium text-red-600 hover:underline disabled:opacity-50"
                      >
                        却下
                      </button>
                    </div>
                  ) : (
                    <Chip tone="neutral">{request.status}</Chip>
                  )}
                </li>
              ))}
            </ul>
          </Card>
        </div>
      ) : null}
    </div>
  );
}
