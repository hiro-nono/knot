"use client";

import { useRouter } from "next/navigation";
import { type FormEvent, useState } from "react";

import { Button } from "@/components/ui/button";
import { FormAlert } from "@/components/ui/form-alert";
import { SelectField } from "@/components/ui/select-field";
import { TextField } from "@/components/ui/text-field";
import { useRegisterAccount } from "@/features/account/hooks";
import { getApiErrorMessage } from "@/lib/api/errors";
import type { AccountType } from "@/types/account";

interface FieldErrors {
  name?: string;
  lastName?: string;
  firstName?: string;
}

// Supabaseでのサインアップ(認証)とは別に、ドメイン上のAccount/Userを
// 作成するための初回登録画面。POST /accountsが未実行だとGET /accounts/meが
// 404になり他の画面が使えないため、認証後は必ずここを経由させる
// ((dashboard)/layout.tsxがGetMe 404を検知してここへ遷移させる)。
export default function OnboardingPage() {
  const router = useRouter();
  const registerAccount = useRegisterAccount();

  const [accountType, setAccountType] = useState<AccountType>("personal");
  const [name, setName] = useState("");
  const [lastName, setLastName] = useState("");
  const [firstName, setFirstName] = useState("");
  const [language, setLanguage] = useState("ja");
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    const errors: FieldErrors = {
      name: accountType === "organization" && !name.trim() ? "組織名を入力してください" : undefined,
      lastName: lastName.trim() ? undefined : "姓を入力してください",
      firstName: firstName.trim() ? undefined : "名を入力してください",
    };
    setFieldErrors(errors);
    if (Object.values(errors).some(Boolean)) {
      return;
    }

    registerAccount.mutate(
      {
        account_type: accountType,
        name: accountType === "organization" ? name : null,
        last_name: lastName,
        first_name: firstName,
        language,
      },
      {
        onSuccess: () => router.push("/"),
      },
    );
  }

  return (
    <div className="flex flex-col gap-6">
      <div>
        <h1 className="text-lg font-semibold text-zinc-900">プロフィールを設定</h1>
        <p className="mt-1 text-sm text-zinc-500">
          利用を開始する前に、基本的なプロフィールを登録してください。
        </p>
      </div>

      <form onSubmit={handleSubmit} noValidate className="flex flex-col gap-4">
        <SelectField
          label="利用形態"
          value={accountType}
          onChange={(e) => setAccountType(e.target.value as AccountType)}
        >
          <option value="personal">個人</option>
          <option value="organization">組織</option>
        </SelectField>

        {accountType === "organization" ? (
          <TextField
            label="組織名"
            value={name}
            onChange={(e) => setName(e.target.value)}
            error={fieldErrors.name}
          />
        ) : null}

        <div className="grid grid-cols-2 gap-3">
          <TextField
            label="姓"
            value={lastName}
            onChange={(e) => setLastName(e.target.value)}
            error={fieldErrors.lastName}
          />
          <TextField
            label="名"
            value={firstName}
            onChange={(e) => setFirstName(e.target.value)}
            error={fieldErrors.firstName}
          />
        </div>

        <SelectField label="言語" value={language} onChange={(e) => setLanguage(e.target.value)}>
          <option value="ja">日本語</option>
          <option value="en">English</option>
        </SelectField>

        {registerAccount.isError ? (
          <FormAlert variant="error">{getApiErrorMessage(registerAccount.error)}</FormAlert>
        ) : null}

        <Button type="submit" isLoading={registerAccount.isPending}>
          はじめる
        </Button>
      </form>
    </div>
  );
}
