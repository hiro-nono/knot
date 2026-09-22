"use client";

import { type FormEvent, useState } from "react";

import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { FormAlert } from "@/components/ui/form-alert";
import { TextField } from "@/components/ui/text-field";
import { useAddRecipient, useRecipients, useRemoveRecipient } from "@/features/information/hooks";
import { UserName } from "@/features/user/components/user-name";
import { ApiError, getApiErrorMessage } from "@/lib/api/errors";

// このInformationを所有するAccountのみがGET /informations/:id/recipientsを
// 呼べる(403 Forbidden)。そのため、ここでの成否を「自分が所有者かどうか」の
// 判定に流用し、所有者以外にはセクションごと表示しない。
export function RecipientsSection({ informationId }: { informationId: string }) {
  const recipients = useRecipients(informationId);
  const addRecipient = useAddRecipient(informationId);
  const removeRecipient = useRemoveRecipient(informationId);

  const [userId, setUserId] = useState("");

  const forbidden = recipients.error instanceof ApiError && recipients.error.isForbidden;

  if (recipients.isLoading || forbidden) {
    return null;
  }

  function handleAdd(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const trimmed = userId.trim();
    if (!trimmed) {
      return;
    }
    addRecipient.mutate(trimmed, { onSuccess: () => setUserId("") });
  }

  return (
    <div>
      <h2 className="mb-2 text-sm font-medium text-zinc-500">共有設定(閲覧できるユーザー)</h2>
      <Card className="flex flex-col gap-4 p-4">
        {recipients.data?.user_ids.length ? (
          <ul className="flex flex-col divide-y divide-zinc-100">
            {recipients.data.user_ids.map((id) => (
              <li key={id} className="flex items-center justify-between gap-2 py-2 first:pt-0 last:pb-0">
                <span className="text-sm">
                  <UserName userId={id} />
                </span>
                <button
                  type="button"
                  disabled={removeRecipient.isPending}
                  onClick={() => removeRecipient.mutate(id)}
                  className="text-sm font-medium text-red-600 hover:underline disabled:opacity-50"
                >
                  削除
                </button>
              </li>
            ))}
          </ul>
        ) : (
          <p className="text-sm text-zinc-500">まだ誰も追加されていません</p>
        )}

        <form onSubmit={handleAdd} className="flex items-end gap-2">
          <div className="flex-1">
            <TextField
              label="ユーザーIDを追加"
              placeholder="user_id"
              value={userId}
              onChange={(e) => setUserId(e.target.value)}
            />
          </div>
          <Button type="submit" isLoading={addRecipient.isPending} className="w-auto px-6">
            追加
          </Button>
        </form>

        {addRecipient.isError ? (
          <FormAlert variant="error">{getApiErrorMessage(addRecipient.error)}</FormAlert>
        ) : null}
        {removeRecipient.isError ? (
          <FormAlert variant="error">{getApiErrorMessage(removeRecipient.error)}</FormAlert>
        ) : null}
      </Card>
    </div>
  );
}
