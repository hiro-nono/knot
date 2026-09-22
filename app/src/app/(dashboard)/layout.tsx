"use client";

import { useRouter } from "next/navigation";
import { useEffect } from "react";

import { AppHeader } from "@/components/layout/app-header";
import { useMyAccount } from "@/features/account/hooks";
import { useSupabaseSession } from "@/hooks/use-supabase-session";
import { ApiError } from "@/lib/api/errors";

// 認証状態の管理は共通hook(useSupabaseSession)に集約し、この配下の画面では
// 個別に実装しない。未認証の場合はサインイン画面へ遷移させる。
//
// 認証済みでもPOST /accountsが未実行(GET /accounts/meが404)の場合は
// ドメイン上のAccount/Userがまだ存在しないため、登録画面(/onboarding)へ誘導する。
export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const { session, isLoading: isSessionLoading } = useSupabaseSession();
  const myAccount = useMyAccount(Boolean(session));

  const accountNotFound = myAccount.error instanceof ApiError && myAccount.error.isNotFound;

  useEffect(() => {
    if (!isSessionLoading && !session) {
      router.replace("/signin");
      return;
    }
    if (accountNotFound) {
      router.replace("/onboarding");
    }
  }, [isSessionLoading, session, accountNotFound, router]);

  if (isSessionLoading || !session || myAccount.isLoading || accountNotFound) {
    return null;
  }

  return (
    <div className="flex flex-1 flex-col">
      <AppHeader />
      <div className="flex flex-1 flex-col">{children}</div>
    </div>
  );
}
