"use client";

import { useRouter } from "next/navigation";
import { type FormEvent, useState } from "react";

import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { FormAlert } from "@/components/ui/form-alert";
import { SelectField } from "@/components/ui/select-field";
import { Spinner } from "@/components/ui/spinner";
import { TextField } from "@/components/ui/text-field";
import { useDeleteMyAccount, useMyAccount, useUpdateMyProfile } from "@/features/account/hooks";
import { useSignOut } from "@/features/auth/hooks";
import { getApiErrorMessage } from "@/lib/api/errors";
import type { AccountView } from "@/types/account";

// ユーザー設定画面。
// プロフィール(姓・名・言語)はGET/PATCH /accounts/meで取得・更新できるため実装する。
//
// Recipientとしての表示に使うPreference(reading_level・verbosity・tone・
// information_priorityなど)は、バックエンドに直接取得・更新するAPIが無く
// (POST /informations/:id/display/chat や /display/comparisons/selection を通じて
// AIが対話の中で間接的に更新するのみ)、現時点ではこの画面から直接編集する手段が無いため
// 対応するAPIが追加されるまで実装しない。
export default function SettingsPage() {
  const myAccount = useMyAccount();

  return (
    <div className="mx-auto flex w-full max-w-lg flex-1 flex-col gap-4 px-4 py-10">
      <h1 className="text-xl font-semibold text-zinc-900">設定</h1>

      {myAccount.isLoading ? (
        <div className="flex justify-center py-10">
          <Spinner />
        </div>
      ) : myAccount.isError || !myAccount.data ? (
        <FormAlert variant="error">プロフィールの取得に失敗しました</FormAlert>
      ) : (
        <>
          <ProfileForm account={myAccount.data} />
          <DangerZone />
        </>
      )}
    </div>
  );
}

function DangerZone() {
  const router = useRouter();
  const deleteAccount = useDeleteMyAccount();
  const signOut = useSignOut();
  const [confirming, setConfirming] = useState(false);

  function handleDelete() {
    deleteAccount.mutate(undefined, {
      onSuccess: () => {
        signOut.mutate(undefined, {
          onSettled: () => router.push("/signin"),
        });
      },
    });
  }

  return (
    <Card className="border-red-200">
      <h2 className="text-sm font-semibold text-red-700">アカウントの削除</h2>
      <p className="mt-1 text-sm text-zinc-500">
        退会するとアカウントは論理削除され、元に戻すことはできません。
      </p>

      {deleteAccount.isError ? (
        <div className="mt-3">
          <FormAlert variant="error">{getApiErrorMessage(deleteAccount.error)}</FormAlert>
        </div>
      ) : null}

      <div className="mt-3">
        {confirming ? (
          <div className="flex items-center gap-3">
            <Button
              type="button"
              onClick={handleDelete}
              isLoading={deleteAccount.isPending || signOut.isPending}
              className="w-auto bg-red-600 px-6 hover:bg-red-700"
            >
              本当に退会する
            </Button>
            <button
              type="button"
              onClick={() => setConfirming(false)}
              className="text-sm font-medium text-zinc-500 hover:underline"
            >
              キャンセル
            </button>
          </div>
        ) : (
          <Button
            type="button"
            onClick={() => setConfirming(true)}
            className="w-auto bg-red-600 px-6 hover:bg-red-700"
          >
            退会する
          </Button>
        )}
      </div>
    </Card>
  );
}

// account.id をkeyにして親から再マウントすることで、フォームの初期値を
// 取得済みのAccountからuseStateの初期値としてそのまま設定できる
// (エフェクトでの同期を避ける)。
function ProfileForm({ account }: { account: AccountView }) {
  const updateProfile = useUpdateMyProfile();

  const [lastName, setLastName] = useState(account.last_name);
  const [firstName, setFirstName] = useState(account.first_name);
  const [language, setLanguage] = useState(account.language);

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    updateProfile.mutate({ last_name: lastName, first_name: firstName, language });
  }

  return (
    <Card>
      <form onSubmit={handleSubmit} className="flex flex-col gap-4">
        <div className="grid grid-cols-2 gap-3">
          <TextField label="姓" value={lastName} onChange={(e) => setLastName(e.target.value)} />
          <TextField label="名" value={firstName} onChange={(e) => setFirstName(e.target.value)} />
        </div>
        <SelectField label="言語" value={language} onChange={(e) => setLanguage(e.target.value)}>
          <option value="ja">日本語</option>
          <option value="en">English</option>
        </SelectField>

        {updateProfile.isError ? (
          <FormAlert variant="error">{getApiErrorMessage(updateProfile.error)}</FormAlert>
        ) : null}
        {updateProfile.isSuccess ? <FormAlert variant="success">更新しました</FormAlert> : null}

        <Button type="submit" isLoading={updateProfile.isPending} className="w-auto self-start px-6">
          保存
        </Button>
      </form>
    </Card>
  );
}
