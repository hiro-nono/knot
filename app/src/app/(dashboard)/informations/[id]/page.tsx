"use client";

import { useParams } from "next/navigation";
import { useEffect, useRef, useState } from "react";

import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Composer } from "@/components/ui/chat/composer";
import { MessageList } from "@/components/ui/chat/message-list";
import type { ChatBubble } from "@/components/ui/chat/types";
import { FormAlert } from "@/components/ui/form-alert";
import { HtmlBody } from "@/components/ui/html-body";
import { Spinner } from "@/components/ui/spinner";
import { RecipientsSection } from "@/features/information/components/recipients-section";
import { PREFERENCE_KEY_OPTIONS } from "@/features/preference/constants";
import { useApplyPreference, useChat, useDisplay, useGenerateComparison } from "@/features/preference/hooks";
import { ResponseForm } from "@/features/response/components/response-form";
import { ResponsesSection } from "@/features/response/components/responses-section";
import { ApiError, getApiErrorMessage } from "@/lib/api/errors";
import type { ComparisonPattern } from "@/types/display";
import type { Message } from "@/types/message";

function newId(): string {
  return `${Date.now()}-${Math.random().toString(36).slice(2)}`;
}

// この情報に対する回答だけを目的とした公開URL(/informations/:id/respond、
// (dashboard)配下ではなくログイン不要)を、受信者に共有できるように提示する。
// access_type・response_policyの組み合わせによっては相手側でログインを
// 求められることもあるが、そこはURL先のページがバックエンドの403を受けて
// 案内するため、ここでは判定せず常に表示する。
function ShareLinkSection({ informationId }: { informationId: string }) {
  const [copied, setCopied] = useState(false);
  const url =
    typeof window !== "undefined"
      ? `${window.location.origin}/informations/${informationId}/respond`
      : `/informations/${informationId}/respond`;

  function handleCopy() {
    navigator.clipboard.writeText(url).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  }

  return (
    <div>
      <h2 className="mb-2 text-sm font-medium text-zinc-500">回答用リンク</h2>
      <Card className="flex flex-col gap-2 p-4 sm:flex-row sm:items-center sm:justify-between">
        <span className="truncate text-sm text-zinc-700">{url}</span>
        <Button type="button" onClick={handleCopy} className="w-auto shrink-0 px-4">
          {copied ? "コピーしました" : "コピー"}
        </Button>
      </Card>
    </div>
  );
}

// 最上部の「あなた向けに最適化された表示」カード。
//
// 受信者がまだ何もPreferenceを記録していない(display.data.has_preferenceが
// false)場合にだけ、アプリ(AI)側が対照的な2パターン(A/B、それぞれ採用した
// 傾向をvalueとして保持)を1回のAI呼び出しでまとめて生成する。既にPreferenceが
// ある受信者には毎回A/Bを見せない。受信者はタブでA/Bの実際の見た目を比較し、
// 気に入った方を「パターンを選択」するだけでよく、選んだ方のvalueがPreference
// として記録される。
function OptimizedDisplayCard({ informationId }: { informationId: string }) {
  const display = useDisplay(informationId);
  const generateComparison = useGenerateComparison(informationId);
  const applyPreference = useApplyPreference(informationId);

  const [comparison, setComparison] = useState<{
    key: string;
    patternA: ComparisonPattern;
    patternB: ComparisonPattern;
  } | null>(null);
  const [activeTab, setActiveTab] = useState<"a" | "b">("a");
  const triggeredRef = useRef(false);

  useEffect(() => {
    if (triggeredRef.current || !display.data || display.data.has_preference) {
      return;
    }
    triggeredRef.current = true;

    const key =
      PREFERENCE_KEY_OPTIONS[Math.floor(Math.random() * PREFERENCE_KEY_OPTIONS.length)].value;
    generateComparison.mutate(
      { key },
      {
        onSuccess: (response) => {
          setComparison({ key: response.key, patternA: response.pattern_a, patternB: response.pattern_b });
          setActiveTab("a");
        },
      },
    );
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [display.data]);

  function handleApply() {
    if (!comparison) {
      return;
    }
    const chosen = activeTab === "a" ? comparison.patternA : comparison.patternB;
    applyPreference.mutate(
      {
        key: comparison.key,
        value: chosen.value,
        appliedDisplay: { ...chosen.display, has_preference: true },
      },
      { onSuccess: () => setComparison(null) },
    );
  }

  const shown = comparison
    ? (activeTab === "a" ? comparison.patternA : comparison.patternB).display
    : display.data;

  return (
    <div className="overflow-hidden rounded-2xl border border-primary-soft bg-gradient-to-b from-primary-soft/60 to-white shadow-sm">
      <div className="flex flex-wrap items-center justify-between gap-2 border-b border-primary-soft/80 px-6 py-4">
        <span className="inline-flex items-center gap-1.5 rounded-full bg-primary-soft px-3 py-1 text-xs font-medium text-primary">
          あなた向けに最適化された表示
        </span>

        {comparison ? (
          <div className="inline-flex rounded-full border border-primary-soft bg-white p-0.5 text-xs font-medium">
            <button
              type="button"
              onClick={() => setActiveTab("a")}
              className={`rounded-full px-3 py-1 transition-colors ${
                activeTab === "a" ? "bg-primary text-primary-foreground" : "text-zinc-500"
              }`}
            >
              パターンA
            </button>
            <button
              type="button"
              onClick={() => setActiveTab("b")}
              className={`rounded-full px-3 py-1 transition-colors ${
                activeTab === "b" ? "bg-primary text-primary-foreground" : "text-zinc-500"
              }`}
            >
              パターンB
            </button>
          </div>
        ) : null}
      </div>

      <div className="px-6 py-5">
        <h1 className="text-2xl font-semibold tracking-tight text-zinc-900">{shown?.title}</h1>
        <HtmlBody className="mt-4">{shown?.body ?? ""}</HtmlBody>

        {comparison ? (
          <div className="mt-4 border-t border-zinc-100 pt-4">
            <Button
              type="button"
              onClick={handleApply}
              isLoading={applyPreference.isPending}
              className="w-auto px-4"
            >
              パターン{activeTab === "a" ? "A" : "B"}を選択
            </Button>
          </div>
        ) : null}

        {applyPreference.isError ? (
          <div className="mt-3">
            <FormAlert variant="error">{getApiErrorMessage(applyPreference.error)}</FormAlert>
          </div>
        ) : null}
      </div>
    </div>
  );
}

export default function InformationDisplayPage() {
  const params = useParams<{ id: string }>();
  const informationId = params.id;

  const display = useDisplay(informationId);
  const chat = useChat(informationId);

  const [history, setHistory] = useState<Message[]>([]);
  const [bubbles, setBubbles] = useState<ChatBubble[]>([]);

  // このChatは資料の情報についての質問(Q&A)であり、表示の見せ方を調整する
  // ものではないため、応答によってcurrentDisplay(表示本文)を書き換えない。
  function handleSend(text: string) {
    setBubbles((prev) => [...prev, { id: newId(), role: "user", content: text }]);

    chat.mutate(
      { messages: history, user_input: text },
      {
        onSuccess: (response) => {
          setHistory(response.messages);
          setBubbles((prev) => [
            ...prev,
            { id: newId(), role: "assistant", content: response.answer },
          ]);
        },
      },
    );
  }

  if (display.isLoading) {
    return (
      <div className="flex flex-1 items-center justify-center">
        <Spinner />
      </div>
    );
  }

  if (display.isError) {
    const error = display.error;
    const message =
      error instanceof ApiError && error.isForbidden
        ? "このInformationを閲覧する権限がありません。"
        : error instanceof ApiError && error.isNotFound
          ? "指定されたInformationが見つかりません。"
          : getApiErrorMessage(error);

    return (
      <div className="mx-auto w-full max-w-lg flex-1 px-4 py-16">
        <FormAlert variant="error">{message}</FormAlert>
      </div>
    );
  }

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-1 flex-col gap-4 px-4 py-6">
      <OptimizedDisplayCard informationId={informationId} />

      <ResponseForm informationId={informationId} />

      <div>
        <h2 className="mb-2 text-sm font-medium text-zinc-500">この資料について質問する(チャット)</h2>
        <Card className="flex min-h-[320px] flex-1 flex-col p-4">
          <MessageList messages={bubbles} isPending={chat.isPending} />

          {chat.isError ? (
            <div className="pb-3">
              <FormAlert variant="error">{getApiErrorMessage(chat.error)}</FormAlert>
            </div>
          ) : null}

          <Composer
            onSend={handleSend}
            disabled={chat.isPending}
            placeholder="この資料の内容について質問してください(例: 集合時間は何時ですか？)"
          />
        </Card>
      </div>

      <ShareLinkSection informationId={informationId} />

      <RecipientsSection informationId={informationId} />
      <ResponsesSection informationId={informationId} />
    </div>
  );
}
