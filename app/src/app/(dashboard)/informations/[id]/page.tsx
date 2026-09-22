"use client";

import { useParams } from "next/navigation";
import { type FormEvent, useState } from "react";

import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Composer } from "@/components/ui/chat/composer";
import { MessageList } from "@/components/ui/chat/message-list";
import type { ChatBubble } from "@/components/ui/chat/types";
import { FormAlert } from "@/components/ui/form-alert";
import { SelectField } from "@/components/ui/select-field";
import { Spinner } from "@/components/ui/spinner";
import { TextField } from "@/components/ui/text-field";
import { RecipientsSection } from "@/features/information/components/recipients-section";
import { CUSTOM_PREFERENCE_KEY, PREFERENCE_KEY_OPTIONS } from "@/features/preference/constants";
import { useChat, useDisplay, usePrepareComparison, useSelectComparison } from "@/features/preference/hooks";
import { describeAssistantTurn } from "@/features/preference/utils";
import { ResponseForm } from "@/features/response/components/response-form";
import { ResponsesSection } from "@/features/response/components/responses-section";
import { ApiError, getApiErrorMessage } from "@/lib/api/errors";
import type { DisplayContent, PrepareComparisonResponse } from "@/types/display";
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

function ComparisonSection({ informationId }: { informationId: string }) {
  const prepareComparison = usePrepareComparison(informationId);
  const selectComparison = useSelectComparison(informationId);

  const [keyOption, setKeyOption] = useState<string>(PREFERENCE_KEY_OPTIONS[0].value);
  const [customKey, setCustomKey] = useState("");
  const [valueA, setValueA] = useState("");
  const [valueB, setValueB] = useState("");
  const [result, setResult] = useState<PrepareComparisonResponse | null>(null);
  const [selectedNotice, setSelectedNotice] = useState(false);

  const resolvedKey = keyOption === CUSTOM_PREFERENCE_KEY ? customKey.trim() : keyOption;
  const canPrepare = Boolean(resolvedKey) && Boolean(valueA.trim()) && Boolean(valueB.trim());

  function handlePrepare(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!canPrepare) {
      return;
    }

    setSelectedNotice(false);
    prepareComparison.mutate(
      { key: resolvedKey, value_a: valueA, value_b: valueB },
      { onSuccess: setResult },
    );
  }

  function handleSelect(selected: "a" | "b") {
    if (!result) {
      return;
    }

    selectComparison.mutate(
      { comparison: result.comparison, selected },
      {
        onSuccess: () => {
          setResult(null);
          setSelectedNotice(true);
        },
      },
    );
  }

  return (
    <div>
      <h2 className="mb-2 text-sm font-medium text-zinc-500">表示を比較して選ぶ(A/B)</h2>
      <Card className="flex flex-col gap-4 p-4">
        <form onSubmit={handlePrepare} className="flex flex-col gap-3">
          <div className="grid gap-3 sm:grid-cols-2">
            <SelectField
              label="比較する項目"
              value={keyOption}
              onChange={(e) => setKeyOption(e.target.value)}
            >
              {PREFERENCE_KEY_OPTIONS.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
              <option value={CUSTOM_PREFERENCE_KEY}>その他(自由入力)</option>
            </SelectField>
            {keyOption === CUSTOM_PREFERENCE_KEY ? (
              <TextField
                label="項目名"
                value={customKey}
                onChange={(e) => setCustomKey(e.target.value)}
                placeholder="例: 文体"
              />
            ) : null}
          </div>

          <div className="grid gap-3 sm:grid-cols-2">
            <TextField
              label="パターンA"
              value={valueA}
              onChange={(e) => setValueA(e.target.value)}
              placeholder="例: 簡潔"
            />
            <TextField
              label="パターンB"
              value={valueB}
              onChange={(e) => setValueB(e.target.value)}
              placeholder="例: 詳細"
            />
          </div>

          {prepareComparison.isError ? (
            <FormAlert variant="error">{getApiErrorMessage(prepareComparison.error)}</FormAlert>
          ) : null}

          <Button
            type="submit"
            isLoading={prepareComparison.isPending}
            disabled={!canPrepare}
            className="w-auto self-start px-6"
          >
            比較を生成
          </Button>
        </form>

        {selectComparison.isError ? (
          <FormAlert variant="error">{getApiErrorMessage(selectComparison.error)}</FormAlert>
        ) : null}
        {selectedNotice ? (
          <FormAlert variant="success">選んだ内容を今後の表示に反映しました。</FormAlert>
        ) : null}

        {result ? (
          <div className="grid gap-3 sm:grid-cols-2">
            {(["a", "b"] as const).map((option) => {
              const content = option === "a" ? result.display_a : result.display_b;
              return (
                <div key={option} className="flex flex-col gap-3 rounded-lg border border-zinc-200 p-4">
                  <div>
                    <p className="mb-1 text-xs font-medium text-zinc-500">
                      パターン{option.toUpperCase()}
                    </p>
                    <h3 className="text-sm font-semibold text-zinc-900">{content.title}</h3>
                    <p className="mt-1 whitespace-pre-wrap text-sm text-zinc-600">{content.body}</p>
                  </div>
                  <Button
                    type="button"
                    onClick={() => handleSelect(option)}
                    isLoading={selectComparison.isPending}
                    className="w-auto self-start px-4"
                  >
                    これに決める
                  </Button>
                </div>
              );
            })}
          </div>
        ) : null}
      </Card>
    </div>
  );
}

export default function InformationDisplayPage() {
  const params = useParams<{ id: string }>();
  const informationId = params.id;

  const display = useDisplay(informationId);
  const chat = useChat(informationId);

  // チャット/A-B比較で表示が更新されるまではuseDisplayの結果をそのまま表示に使う。
  const [overrideDisplay, setOverrideDisplay] = useState<DisplayContent | null>(null);
  const [history, setHistory] = useState<Message[]>([]);
  const [bubbles, setBubbles] = useState<ChatBubble[]>([]);

  const currentDisplay = overrideDisplay ?? display.data ?? null;

  function handleSend(text: string) {
    setBubbles((prev) => [...prev, { id: newId(), role: "user", content: text }]);

    chat.mutate(
      { messages: history, user_input: text },
      {
        onSuccess: (response) => {
          setHistory(response.messages);
          setOverrideDisplay(response.display);
          setBubbles((prev) => [
            ...prev,
            { id: newId(), role: "assistant", content: describeAssistantTurn(response) },
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
      <Card>
        <h1 className="text-2xl font-semibold text-zinc-900">{currentDisplay?.title}</h1>
        <p className="mt-3 whitespace-pre-wrap text-sm leading-relaxed text-zinc-700">
          {currentDisplay?.body}
        </p>
      </Card>

      <ResponseForm informationId={informationId} />

      <div>
        <h2 className="mb-2 text-sm font-medium text-zinc-500">表示の調整(チャット)</h2>
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
            placeholder="表示について要望を伝えてください(例: もっと簡潔に)"
          />
        </Card>
      </div>

      <ComparisonSection informationId={informationId} />

      <ShareLinkSection informationId={informationId} />

      <RecipientsSection informationId={informationId} />
      <ResponsesSection informationId={informationId} />
    </div>
  );
}
