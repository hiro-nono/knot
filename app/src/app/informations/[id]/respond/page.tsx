"use client";

import Link from "next/link";
import { useParams } from "next/navigation";

import { Card } from "@/components/ui/card";
import { FormAlert } from "@/components/ui/form-alert";
import { Spinner } from "@/components/ui/spinner";
import { useSources } from "@/features/information/hooks";
import { ResponseForm } from "@/features/response/components/response-form";
import { ApiError, getApiErrorMessage } from "@/lib/api/errors";

// PUBLIC+ANONYMOUSなInformationにも未ログインで到達できる、(dashboard)配下ではない
// 公開ページ。作成者向けの完全な表示(/informations/:id、要ログイン)とは別に、
// 回答だけを目的とした受信者向けの入り口として存在する。
//
// GET /informations/:id/sources はSubmitResponseと同じアクセス制御
// (canRespondToInformation)なので、認証の有無に関わらずここで呼んでよい。
// access_type・response_policyの組み合わせによっては認証が必要になるため、
// その場合はバックエンドが返す403を受けてログインを促す。
export default function RespondPage() {
  const params = useParams<{ id: string }>();
  const informationId = params.id;

  const sources = useSources(informationId);

  if (sources.isLoading) {
    return (
      <div className="flex flex-1 items-center justify-center py-16">
        <Spinner />
      </div>
    );
  }

  if (sources.isError) {
    const error = sources.error;
    const forbidden = error instanceof ApiError && error.isForbidden;
    const message = forbidden
      ? "この情報に回答するにはログインが必要です。"
      : error instanceof ApiError && error.isNotFound
        ? "指定された情報が見つかりません。"
        : getApiErrorMessage(error);

    return (
      <div className="mx-auto w-full max-w-lg flex-1 px-4 py-16">
        <FormAlert variant="error">{message}</FormAlert>
        {forbidden ? (
          <div className="mt-4 text-center text-sm">
            <Link
              href={`/signin?next=${encodeURIComponent(`/informations/${informationId}/respond`)}`}
              className="font-medium text-primary hover:underline"
            >
              ログインする
            </Link>
          </div>
        ) : null}
      </div>
    );
  }

  return (
    <div className="mx-auto flex w-full max-w-lg flex-1 flex-col gap-4 px-4 py-10">
      <Card>
        <h1 className="text-xl font-semibold text-zinc-900">アンケートに回答</h1>
      </Card>
      <ResponseForm informationId={informationId} />
    </div>
  );
}
