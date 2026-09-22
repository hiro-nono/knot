"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { type FormEvent, Suspense, useState } from "react";

import { Button } from "@/components/ui/button";
import { FormAlert } from "@/components/ui/form-alert";
import { TextField } from "@/components/ui/text-field";
import { getAuthErrorMessage } from "@/features/auth/errors";
import { useSignIn } from "@/features/auth/hooks";
import { validateEmail, validateRequired } from "@/features/auth/validation";

interface FieldErrors {
  email?: string;
  password?: string;
}

function SignInForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const signIn = useSignIn();

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});

  const confirmFailed = searchParams.get("error") === "confirm_failed";
  const rawErrorMessage = signIn.error instanceof Error ? signIn.error.message : undefined;
  const isEmailNotConfirmed = rawErrorMessage === "Email not confirmed";

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const errors: FieldErrors = {
      email: validateEmail(email),
      password: validateRequired(password, "パスワードを入力してください"),
    };
    setFieldErrors(errors);
    if (Object.values(errors).some(Boolean)) {
      return;
    }

    signIn.mutate(
      { email, password },
      {
        onSuccess: () => {
          router.push(searchParams.get("next") ?? "/");
        },
      },
    );
  }

  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-lg font-semibold text-zinc-900">サインイン</h1>

      {confirmFailed ? (
        <FormAlert variant="error">確認リンクが無効か、有効期限が切れています</FormAlert>
      ) : null}

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
          autoComplete="current-password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          error={fieldErrors.password}
        />

        {signIn.isError ? (
          <FormAlert variant="error">
            {getAuthErrorMessage(signIn.error)}
            {isEmailNotConfirmed ? (
              <>
                {" "}
                <Link
                  href={`/verify-email?type=signup&email=${encodeURIComponent(email)}`}
                  className="font-medium underline"
                >
                  確認メールを再送信する
                </Link>
              </>
            ) : null}
          </FormAlert>
        ) : null}

        <Button type="submit" isLoading={signIn.isPending}>
          サインイン
        </Button>
      </form>

      <div className="flex flex-col gap-2 text-sm text-zinc-600">
        <Link href="/reset-password" className="font-medium text-primary hover:underline">
          パスワードをお忘れの場合
        </Link>
        <p>
          アカウントをお持ちでない場合は{" "}
          <Link href="/signup" className="font-medium text-primary hover:underline">
            登録
          </Link>
        </p>
      </div>
    </div>
  );
}

export default function SignInPage() {
  return (
    <Suspense fallback={null}>
      <SignInForm />
    </Suspense>
  );
}
