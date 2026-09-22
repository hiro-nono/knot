"use client";

import { Card } from "@/components/ui/card";
import { FormAlert } from "@/components/ui/form-alert";
import { useResponses } from "@/features/response/hooks";
import { ApiError, getApiErrorMessage } from "@/lib/api/errors";

function shortId(id: string): string {
  return id.slice(0, 8);
}

// GET /informations/:id/responses もこのInformationの所有者専用のため、
// 403を「所有者ではない」判定に流用してセクションごと非表示にする。
//
// source_id/option_idの人間向けラベル(key・選択肢の文言)を取得するAPIが
// 現状無いため、IDの先頭のみを表示する簡易表示に留めている。
export function ResponsesSection({ informationId }: { informationId: string }) {
  const responses = useResponses(informationId);

  const forbidden = responses.error instanceof ApiError && responses.error.isForbidden;

  if (responses.isLoading || forbidden) {
    return null;
  }

  return (
    <div>
      <h2 className="mb-2 text-sm font-medium text-zinc-500">
        寄せられた回答{responses.data ? `(${responses.data.length}件)` : ""}
      </h2>
      <Card className="flex flex-col gap-3 p-4">
        {responses.isError ? (
          <FormAlert variant="error">{getApiErrorMessage(responses.error)}</FormAlert>
        ) : responses.data?.length ? (
          <ul className="flex flex-col divide-y divide-zinc-100">
            {responses.data.map((response) => (
              <li key={response.id} className="py-3 first:pt-0 last:pb-0">
                <p className="mb-1 text-xs font-medium text-zinc-500">
                  {response.user_id ? `ユーザー: ${shortId(response.user_id)}` : "匿名"}
                </p>
                <ul className="flex flex-col gap-1">
                  {response.items.map((item) => (
                    <li key={item.source_id} className="text-sm text-zinc-700">
                      <span className="text-zinc-400">{shortId(item.source_id)}: </span>
                      {item.value ?? (item.option_id ? shortId(item.option_id) : "-")}
                    </li>
                  ))}
                </ul>
              </li>
            ))}
          </ul>
        ) : (
          <p className="text-sm text-zinc-500">まだ回答がありません</p>
        )}
      </Card>
    </div>
  );
}
