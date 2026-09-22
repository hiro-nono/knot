"use client";

import Link from "next/link";
import { type FormEvent, useState } from "react";

import { Button } from "@/components/ui/button";
import { FormAlert } from "@/components/ui/form-alert";
import { TextField } from "@/components/ui/text-field";
import { getAuthErrorMessage } from "@/features/auth/errors";
import { useRequestPasswordReset } from "@/features/auth/hooks";
import { validateEmail } from "@/features/auth/validation";

export default function ResetPasswordRequestPage() {
  const requestReset = useRequestPasswordReset();

  const [email, setEmail] = useState("");
  const [emailError, setEmailError] = useState<string>();

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const error = validateEmail(email);
    setEmailError(error);
    if (error) {
      return;
    }

    requestReset.mutate({ email });
  }

  if (requestReset.isSuccess) {
    return (
      <div className="flex flex-col gap-4">
        <h1 className="text-lg font-semibold text-zinc-900">メールを送信しました</h1>
        <FormAlert variant="success">
          {email} 宛にパスワード再設定用のリンクを送信しました。メールをご確認ください。
        </FormAlert>
        <Link href="/signin" className="text-sm font-medium text-primary hover:underline">
          サインインに戻る
        </Link>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-lg font-semibold text-zinc-900">パスワード再設定</h1>
      <p className="text-sm text-zinc-600">
        登録済みのメールアドレスに再設定用のリンクを送信します。
      </p>

      <form onSubmit={handleSubmit} noValidate className="flex flex-col gap-4">
        <TextField
          label="メールアドレス"
          type="email"
          autoComplete="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          error={emailError}
        />

        {requestReset.isError ? (
          <FormAlert variant="error">{getAuthErrorMessage(requestReset.error)}</FormAlert>
        ) : null}

        <Button type="submit" isLoading={requestReset.isPending}>
          再設定メールを送信
        </Button>
      </form>

      <Link href="/signin" className="text-sm font-medium text-primary hover:underline">
        サインインに戻る
      </Link>
    </div>
  );
}
