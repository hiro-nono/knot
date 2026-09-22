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

  // onboarding完了直後は、POST /accountsのonSuccessでGET /accounts/meの
  // キャッシュをinvalidateしてから"/"へ遷移してくる。その際、再取得
  // (isFetching)が終わるまではキャッシュに古い404エラーが残ったままになるため、
  // isFetching中はそのエラーを「未登録」と確定させず、再取得の結果を待つ。
  // これを怠ると、登録直後でも古い404を拾って/onboardingへ引き戻してしまう。
  const isResolvingAccount = myAccount.isLoading || (myAccount.isFetching && myAccount.isError);
  const accountNotFound =
    !isResolvingAccount && myAccount.error instanceof ApiError && myAccount.error.isNotFound;

  useEffect(() => {
    if (!isSessionLoading && !session) {
      router.replace("/signin");
      return;
    }
    if (accountNotFound) {
      router.replace("/onboarding");
    }
  }, [isSessionLoading, session, accountNotFound, router]);

  if (isSessionLoading || !session || isResolvingAccount || accountNotFound) {
    return null;
  }

  return (
    <div className="flex flex-1 flex-col">
      <AppHeader />
      <div className="flex flex-1 flex-col">{children}</div>
    </div>
  );
}
