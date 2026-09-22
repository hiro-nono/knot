"use client";

import Link from "next/link";
import { type FormEvent, useState } from "react";

import { Button } from "@/components/ui/button";
import { FormAlert } from "@/components/ui/form-alert";
import { TextField } from "@/components/ui/text-field";
import { getAuthErrorMessage } from "@/features/auth/errors";
import { useChangeEmail } from "@/features/auth/hooks";
import { validateEmail } from "@/features/auth/validation";
import { useSupabaseSession } from "@/hooks/use-supabase-session";

export default function ChangeEmailPage() {
  const { session, isLoading } = useSupabaseSession();
  const changeEmail = useChangeEmail();

  const [email, setEmail] = useState("");
  const [emailError, setEmailError] = useState<string>();

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const error = validateEmail(email);
    setEmailError(error);
    if (error) {
      return;
    }

    changeEmail.mutate({ email });
  }

  if (isLoading) {
    return null;
  }

  if (!session) {
    return (
      <div className="flex flex-col gap-4">
        <h1 className="text-lg font-semibold text-zinc-900">ログインが必要です</h1>
        <Link href="/signin" className="text-sm font-medium text-primary hover:underline">
          サインインへ
        </Link>
      </div>
    );
  }

  if (changeEmail.isSuccess) {
    return (
      <div className="flex flex-col gap-4">
        <h1 className="text-lg font-semibold text-zinc-900">確認メールを送信しました</h1>
        <FormAlert variant="success">
          {email}{" "}
          宛に確認メールを送信しました。リンクをクリックして変更を完了してください。
        </FormAlert>
        <Link
          href={`/verify-email?type=email_change&email=${encodeURIComponent(email)}`}
          className="text-sm font-medium text-primary hover:underline"
        >
          メールが届かない場合はこちら
        </Link>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-6">
      <Link href="/" className="text-sm font-medium text-primary hover:underline">
        ← ホームへ戻る
      </Link>
      <h1 className="text-lg font-semibold text-zinc-900">メールアドレスの変更</h1>
      <p className="text-sm text-zinc-600">
        新しいメールアドレス宛に確認メールを送信します。リンクをクリックするまで変更は反映されません。
      </p>

      <form onSubmit={handleSubmit} noValidate className="flex flex-col gap-4">
        <TextField
          label="新しいメールアドレス"
          type="email"
          autoComplete="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          error={emailError}
        />

        {changeEmail.isError ? (
          <FormAlert variant="error">{getAuthErrorMessage(changeEmail.error)}</FormAlert>
        ) : null}

        <Button type="submit" isLoading={changeEmail.isPending}>
          確認メールを送信
        </Button>
      </form>
    </div>
  );
}
