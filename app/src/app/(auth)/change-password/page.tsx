"use client";

import Link from "next/link";
import { type FormEvent, useState } from "react";

import { Button } from "@/components/ui/button";
import { FormAlert } from "@/components/ui/form-alert";
import { TextField } from "@/components/ui/text-field";
import { getAuthErrorMessage } from "@/features/auth/errors";
import { useChangePassword } from "@/features/auth/hooks";
import { validatePassword, validatePasswordConfirmation } from "@/features/auth/validation";
import { useSupabaseSession } from "@/hooks/use-supabase-session";

interface FieldErrors {
  password?: string;
  confirmPassword?: string;
}

export default function ChangePasswordPage() {
  const { session, isLoading } = useSupabaseSession();
  const changePassword = useChangePassword();

  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const errors: FieldErrors = {
      password: validatePassword(password),
      confirmPassword: validatePasswordConfirmation(password, confirmPassword),
    };
    setFieldErrors(errors);
    if (Object.values(errors).some(Boolean)) {
      return;
    }

    changePassword.mutate(
      { password },
      {
        onSuccess: () => {
          setPassword("");
          setConfirmPassword("");
        },
      },
    );
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

  return (
    <div className="flex flex-col gap-6">
      <Link href="/" className="text-sm font-medium text-primary hover:underline">
        ← ホームへ戻る
      </Link>
      <h1 className="text-lg font-semibold text-zinc-900">パスワードの変更</h1>

      <form onSubmit={handleSubmit} noValidate className="flex flex-col gap-4">
        <TextField
          label="新しいパスワード"
          type="password"
          autoComplete="new-password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          error={fieldErrors.password}
        />
        <TextField
          label="新しいパスワード(確認)"
          type="password"
          autoComplete="new-password"
          value={confirmPassword}
          onChange={(e) => setConfirmPassword(e.target.value)}
          error={fieldErrors.confirmPassword}
        />

        {changePassword.isError ? (
          <FormAlert variant="error">{getAuthErrorMessage(changePassword.error)}</FormAlert>
        ) : null}
        {changePassword.isSuccess ? (
          <FormAlert variant="success">パスワードを変更しました</FormAlert>
        ) : null}

        <Button type="submit" isLoading={changePassword.isPending}>
          パスワードを変更
        </Button>
      </form>
    </div>
  );
}
