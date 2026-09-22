"use client";

import Link from "next/link";

import { Card } from "@/components/ui/card";
import { useMyAccount } from "@/features/account/hooks";
import { useMyInformations } from "@/features/information/hooks";

const ACCESS_TYPE_LABEL: Record<string, string> = {
  public: "リンクを知っている全員",
  restricted: "限定公開",
};

export default function Home() {
  const { data: account } = useMyAccount();
  const myInformations = useMyInformations();

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-1 flex-col gap-6 px-4 py-10">
      <div>
        <h1 className="text-2xl font-semibold text-zinc-900">
          {account ? `こんにちは、${account.first_name}さん` : "ようこそ"}
        </h1>
        <p className="mt-1 text-sm text-zinc-500">
          伝えたい情報をチャットで整理し、受け手にあわせて最適な形で届けます。
        </p>
      </div>

      <Link href="/informations/new">
        <Card className="flex items-center justify-between gap-4 transition-shadow hover:shadow-md">
          <div>
            <h2 className="text-base font-semibold text-zinc-900">新しい情報を作成する</h2>
            <p className="mt-1 text-sm text-zinc-500">
              チャットで話すだけで、AIが情報を整理・確定します。
            </p>
          </div>
          <span className="text-2xl text-primary" aria-hidden="true">
            →
          </span>
        </Card>
      </Link>

      {myInformations.data?.length ? (
        <div>
          <h2 className="mb-2 text-sm font-medium text-zinc-500">作成した情報</h2>
          <Card className="flex flex-col gap-1 p-2">
            <ul className="flex flex-col divide-y divide-zinc-100">
              {myInformations.data.map((information) => (
                <li key={information.id}>
                  <Link
                    href={`/informations/${information.id}`}
                    className="flex items-center justify-between gap-2 rounded-lg px-2 py-2.5 hover:bg-zinc-50"
                  >
                    <span className="text-sm text-zinc-900">{information.title}</span>
                    <span className="text-xs text-zinc-400">
                      {ACCESS_TYPE_LABEL[information.access_type] ?? information.access_type}
                    </span>
                  </Link>
                </li>
              ))}
            </ul>
          </Card>
        </div>
      ) : null}

      <div className="grid gap-4 sm:grid-cols-2">
        <Link href="/account/members">
          <Card className="transition-shadow hover:shadow-md">
            <h2 className="text-sm font-semibold text-zinc-900">メンバー管理</h2>
            <p className="mt-1 text-sm text-zinc-500">組織のメンバーや権限を管理します。</p>
          </Card>
        </Link>
        <Link href="/settings">
          <Card className="transition-shadow hover:shadow-md">
            <h2 className="text-sm font-semibold text-zinc-900">設定</h2>
            <p className="mt-1 text-sm text-zinc-500">プロフィールを編集します。</p>
          </Card>
        </Link>
      </div>
    </div>
  );
}
