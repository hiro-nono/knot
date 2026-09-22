"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { type FormEvent, useState } from "react";

import { Button } from "@/components/ui/button";
import { FormAlert } from "@/components/ui/form-alert";
import { TextField } from "@/components/ui/text-field";
import { getAuthErrorMessage } from "@/features/auth/errors";
import { useSignUp } from "@/features/auth/hooks";
import {
  validateEmail,
  validatePassword,
  validatePasswordConfirmation,
} from "@/features/auth/validation";

interface FieldErrors {
  email?: string;
  password?: string;
  confirmPassword?: string;
}

export default function SignUpPage() {
  const router = useRouter();
  const signUp = useSignUp();

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const errors: FieldErrors = {
      email: validateEmail(email),
      password: validatePassword(password),
      confirmPassword: validatePasswordConfirmation(password, confirmPassword),
    };
    setFieldErrors(errors);
    if (Object.values(errors).some(Boolean)) {
      return;
    }

    signUp.mutate(
      { email, password },
      {
        onSuccess: () => {
          router.push(`/verify-email?type=signup&email=${encodeURIComponent(email)}`);
        },
      },
    );
  }

  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-lg font-semibold text-zinc-900">アカウント登録</h1>

      <form onSubmit={handleSubmit} noValidate className="flex flex-col gap-4">
        <TextField
          label="メールアドレス"
          type="email"
          autoComplete="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          error={fieldErrors.email}
        />
        <TextField
          label="パスワード"
          type="password"
          autoComplete="new-password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          error={fieldErrors.password}
        />
        <TextField
          label="パスワード(確認)"
          type="password"
          autoComplete="new-password"
          value={confirmPassword}
          onChange={(e) => setConfirmPassword(e.target.value)}
          error={fieldErrors.confirmPassword}
        />

        {signUp.isError ? <FormAlert variant="error">{getAuthErrorMessage(signUp.error)}</FormAlert> : null}

        <Button type="submit" isLoading={signUp.isPending}>
          登録する
        </Button>
      </form>

      <p className="text-sm text-zinc-600">
        アカウントをお持ちの場合は{" "}
        <Link href="/signin" className="font-medium text-primary hover:underline">
          サインイン
        </Link>
      </p>
    </div>
  );
}
