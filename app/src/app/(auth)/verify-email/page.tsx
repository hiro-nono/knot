"use client";

import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { type FormEvent, Suspense, useState } from "react";

import { Button } from "@/components/ui/button";
import { FormAlert } from "@/components/ui/form-alert";
import { TextField } from "@/components/ui/text-field";
import { getAuthErrorMessage } from "@/features/auth/errors";
import { useResendVerificationEmail } from "@/features/auth/hooks";
import { validateEmail } from "@/features/auth/validation";
import type { ResendVerificationEmailInput } from "@/features/auth/api";

function isResendType(value: string | null): value is ResendVerificationEmailInput["type"] {
  return value === "signup" || value === "email_change";
}

function VerifyEmailContent() {
  const searchParams = useSearchParams();
  const resend = useResendVerificationEmail();

  const typeParam = searchParams.get("type");
  const type: ResendVerificationEmailInput["type"] = isResendType(typeParam) ? typeParam : "signup";
  const [email, setEmail] = useState(searchParams.get("email") ?? "");
  const [emailError, setEmailError] = useState<string>();

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const error = validateEmail(email);
    setEmailError(error);
    if (error) {
      return;
    }

    resend.mutate({ type, email });
  }

  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-lg font-semibold text-zinc-900">メールを確認してください</h1>
      <p className="text-sm text-zinc-600">
        {type === "email_change"
          ? "新しいメールアドレス宛に確認リンクを送信しています。届かない場合は再送信してください。"
          : "登録されたメールアドレス宛に確認リンクを送信しています。届かない場合は再送信してください。"}
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

        {resend.isError ? <FormAlert variant="error">{getAuthErrorMessage(resend.error)}</FormAlert> : null}
        {resend.isSuccess ? <FormAlert variant="success">確認メールを再送信しました</FormAlert> : null}

        <Button type="submit" isLoading={resend.isPending}>
          確認メールを再送信
        </Button>
      </form>

      <Link href="/signin" className="text-sm font-medium text-primary hover:underline">
        サインインに戻る
      </Link>
    </div>
  );
}

export default function VerifyEmailPage() {
  return (
    <Suspense fallback={null}>
      <VerifyEmailContent />
    </Suspense>
  );
}
