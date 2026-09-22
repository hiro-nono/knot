"use client";

import { useUserProfile } from "@/features/user/hooks";

// メンバー一覧・recipient一覧などはuser_id(UUID)しか返さないため、
// GET /users/:id で解決した氏名を表示する。取得できるまで/失敗した場合はIDを表示する。
export function UserName({ userId }: { userId: string }) {
  const { data, isLoading } = useUserProfile(userId);

  if (isLoading || !data) {
    return <span className="text-zinc-900">{userId}</span>;
  }

  return (
    <span className="text-zinc-900">
      {data.last_name} {data.first_name}
    </span>
  );
}
