"use client";

import { useState } from "react";

import { Card } from "@/components/ui/card";
import { Chip, type ChipTone } from "@/components/ui/chip";
import { FormAlert } from "@/components/ui/form-alert";
import { Spinner } from "@/components/ui/spinner";
import { useAccountsByStatus, useUpdateAccountStatus } from "@/features/account/hooks";
import { ApiError, getApiErrorMessage } from "@/lib/api/errors";
import type { AccountStatus, AccountStatusAction } from "@/types/account";

const STATUS_TABS: { value: AccountStatus; label: string }[] = [
  { value: "active", label: "有効" },
  { value: "frozen", label: "凍結" },
  { value: "suspended", label: "一時停止" },
  { value: "banned", label: "アカウント停止" },
];

const STATUS_TONE: Record<AccountStatus, ChipTone> = {
  active: "success",
  frozen: "neutral",
  suspended: "warning",
  banned: "warning",
  withdrawn: "neutral",
};

const ACTIONS_FOR_STATUS: Record<AccountStatus, { action: AccountStatusAction; label: string }[]> = {
  active: [
    { action: "freeze", label: "凍結する" },
    { action: "suspend", label: "一時停止する" },
    { action: "ban", label: "アカウント停止する" },
  ],
  frozen: [{ action: "unfreeze", label: "凍結を解除する" }],
  suspended: [{ action: "reactivate", label: "復帰させる" }],
  banned: [{ action: "reactivate", label: "復帰させる" }],
  withdrawn: [],
};

// GET /accounts?status=... / PATCH /accounts/:id/status はadmin(AccountRole)専用。
// 権限が無い場合は一覧取得が403になるので、そのままエラー表示する
// (organization内のowner/adminとは別の、プラットフォーム全体の管理者フラグ)。
export default function AdminAccountsPage() {
  const [status, setStatus] = useState<AccountStatus>("active");
  const accounts = useAccountsByStatus(status);
  const updateStatus = useUpdateAccountStatus();

  const forbidden = accounts.error instanceof ApiError && accounts.error.isForbidden;

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-1 flex-col gap-4 px-4 py-10">
      <h1 className="text-xl font-semibold text-zinc-900">アカウント管理</h1>

      <div className="flex gap-1">
        {STATUS_TABS.map((tab) => (
          <button
            key={tab.value}
            type="button"
            onClick={() => setStatus(tab.value)}
            className={`rounded-full px-3 py-1.5 text-sm font-medium transition-colors ${
              status === tab.value
                ? "bg-primary-soft text-primary-hover"
                : "text-zinc-600 hover:bg-zinc-100"
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {forbidden ? (
        <FormAlert variant="error">この画面を利用する権限がありません(管理者のみ利用できます)</FormAlert>
      ) : accounts.isLoading ? (
        <div className="flex justify-center py-10">
          <Spinner />
        </div>
      ) : (
        <Card className="flex flex-col gap-3 p-4">
          {updateStatus.isError ? (
            <FormAlert variant="error">{getApiErrorMessage(updateStatus.error)}</FormAlert>
          ) : null}

          {accounts.data?.length ? (
            <ul className="flex flex-col divide-y divide-zinc-100">
              {accounts.data.map((account) => (
                <li
                  key={account.id}
                  className="flex flex-wrap items-center justify-between gap-2 py-3 first:pt-0 last:pb-0"
                >
                  <div className="flex flex-col gap-1">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium text-zinc-900">
                        {account.name ?? `${account.last_name} ${account.first_name}`}
                      </span>
                      <Chip tone={STATUS_TONE[account.status]}>{account.status}</Chip>
                    </div>
                    <span className="text-xs text-zinc-400">{account.id}</span>
                  </div>
                  <div className="flex gap-3">
                    {ACTIONS_FOR_STATUS[account.status].map(({ action, label }) => (
                      <button
                        key={action}
                        type="button"
                        disabled={updateStatus.isPending}
                        onClick={() => updateStatus.mutate({ accountId: account.id, action })}
                        className="text-sm font-medium text-primary hover:underline disabled:opacity-50"
                      >
                        {label}
                      </button>
                    ))}
                  </div>
                </li>
              ))}
            </ul>
          ) : (
            <p className="py-4 text-center text-sm text-zinc-500">該当するアカウントはありません</p>
          )}
        </Card>
      )}
    </div>
  );
}
