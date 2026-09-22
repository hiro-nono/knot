"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { type FormEvent, useState } from "react";

import { Button } from "@/components/ui/button";
import { FormAlert } from "@/components/ui/form-alert";
import { TextField } from "@/components/ui/text-field";
import { getAuthErrorMessage } from "@/features/auth/errors";
import { useResetPassword, useSignOut } from "@/features/auth/hooks";
import { validatePassword, validatePasswordConfirmation } from "@/features/auth/validation";
import { useSupabaseSession } from "@/hooks/use-supabase-session";

interface FieldErrors {
  password?: string;
  confirmPassword?: string;
}

export default function ResetPasswordConfirmPage() {
  const router = useRouter();
  const { session, isLoading } = useSupabaseSession();
  const resetPassword = useResetPassword();
  const signOut = useSignOut();

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

    resetPassword.mutate(
      { password },
      {
        onSuccess: () => {
          // メールリンク経由の一時セッションを終了し、新しいパスワードで
          // 改めてサインインしてもらう。
          signOut.mutate(undefined, {
            onSettled: () => {
              router.push("/signin");
            },
          });
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
        <h1 className="text-lg font-semibold text-zinc-900">リンクが無効です</h1>
        <FormAlert variant="error">
          パスワード再設定用のリンクが無効か、有効期限が切れています。
        </FormAlert>
        <Link href="/reset-password" className="text-sm font-medium text-primary hover:underline">
          再度メールを送信する
        </Link>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-lg font-semibold text-zinc-900">新しいパスワードを設定</h1>

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

        {resetPassword.isError ? (
          <FormAlert variant="error">{getAuthErrorMessage(resetPassword.error)}</FormAlert>
        ) : null}

        <Button type="submit" isLoading={resetPassword.isPending || signOut.isPending}>
          パスワードを更新
        </Button>
      </form>
    </div>
  );
}
