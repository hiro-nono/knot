"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Chip, type ChipTone } from "@/components/ui/chip";
import { Composer } from "@/components/ui/chat/composer";
import { MessageList } from "@/components/ui/chat/message-list";
import type { ChatBubble } from "@/components/ui/chat/types";
import { FormAlert } from "@/components/ui/form-alert";
import { SelectField } from "@/components/ui/select-field";
import { useMyAccount } from "@/features/account/hooks";
import { addRecipient } from "@/features/information/api";
import { useProcessInformation } from "@/features/information/hooks";
import { describeAssistantTurn } from "@/features/information/utils";
import { getApiErrorMessage } from "@/lib/api/errors";
import type {
  InformationAccessType,
  InformationResponsePolicy,
  StructuredSource,
} from "@/types/information";
import type { Message } from "@/types/message";

function newId(): string {
  return `${Date.now()}-${Math.random().toString(36).slice(2)}`;
}

const STATUS_TONE: Record<string, ChipTone> = {
  confirmed: "success",
  undecided: "warning",
  unknown: "neutral",
};

const STATUS_LABEL: Record<string, string> = {
  confirmed: "確定",
  undecided: "未確定",
  unknown: "不明",
};

function SourceCard({ source }: { source: StructuredSource }) {
  return (
    <div className="rounded-lg border border-zinc-200 p-3">
      <div className="mb-1.5 flex items-center justify-between gap-2">
        <span className="text-sm font-medium text-zinc-900">{source.key}</span>
        <Chip tone={STATUS_TONE[source.status] ?? "neutral"}>
          {STATUS_LABEL[source.status] ?? source.status}
        </Chip>
      </div>
      <p className="text-sm text-zinc-600">{source.value}</p>
      {source.options && source.options.length > 0 ? (
        <div className="mt-2 flex flex-wrap gap-1.5">
          {source.options.map((option) => (
            <Chip key={option.value} tone="primary">
              {option.value}
            </Chip>
          ))}
        </div>
      ) : null}
    </div>
  );
}

export default function NewInformationPage() {
  const router = useRouter();
  const { data: myAccount } = useMyAccount();
  const processInformation = useProcessInformation();

  const [accessType, setAccessType] = useState<InformationAccessType>("restricted");
  const [responsePolicy, setResponsePolicy] = useState<InformationResponsePolicy>("authenticated");
  const [history, setHistory] = useState<Message[]>([]);
  const [bubbles, setBubbles] = useState<ChatBubble[]>([]);
  const [sources, setSources] = useState<StructuredSource[]>([]);
  const [savedInformationId, setSavedInformationId] = useState<string | null>(null);

  const started = history.length > 0;

  function handleSend(text: string) {
    setBubbles((prev) => [...prev, { id: newId(), role: "user", content: text }]);

    processInformation.mutate(
      {
        access_type: accessType,
        response_policy: responsePolicy,
        messages: history,
        user_input: text,
      },
      {
        onSuccess: async (response) => {
          setHistory(response.messages);
          setSources(response.structured?.sources ?? []);
          setBubbles((prev) => [
            ...prev,
            { id: newId(), role: "assistant", content: describeAssistantTurn(response) },
          ]);

          // すべての項目が確定した時点でバックエンドが既にInformationを
          // 保存済みのため、ここではまだ遷移せず「保存する」ボタンを
          // 押せる状態にするだけに留める。実際にいつ遷移するかは
          // 発信者自身が内容を確認してから選べるようにする。
          if (response.confirmed && response.information_id) {
            if (accessType === "restricted" && myAccount?.user_id) {
              try {
                await addRecipient(response.information_id, myAccount.user_id);
              } catch {
                // 自分をrecipientに追加できなくても保存自体は完了しているので続行する。
              }
            }
            setSavedInformationId(response.information_id);
          }
        },
      },
    );
  }

  function handleConfirmSave() {
    if (savedInformationId) {
      router.push(`/informations/${savedInformationId}`);
    }
  }

  return (
    <div className="mx-auto flex w-full max-w-3xl flex-1 flex-col gap-4 px-4 py-6">
      <div>
        <h1 className="text-xl font-semibold text-zinc-900">新しい情報を作成</h1>
        <p className="text-sm text-zinc-500">
          チャットで内容を伝えると、AIが情報を整理します。
        </p>
      </div>

      <div className="grid grid-cols-2 gap-3 sm:max-w-md">
        <SelectField
          label="公開範囲"
          value={accessType}
          disabled={started}
          onChange={(e) => setAccessType(e.target.value as InformationAccessType)}
        >
          <option value="restricted">限定公開(招待した人のみ)</option>
          <option value="public">リンクを知っている全員</option>
        </SelectField>
        <SelectField
          label="回答の受付"
          value={responsePolicy}
          disabled={started || accessType === "restricted"}
          onChange={(e) => setResponsePolicy(e.target.value as InformationResponsePolicy)}
        >
          <option value="authenticated">ログイン済みのみ</option>
          <option value="anonymous">未ログインも可</option>
        </SelectField>
      </div>

      <div className="grid flex-1 grid-cols-1 gap-4 lg:grid-cols-[1fr_300px]">
        <Card className="flex min-h-[420px] flex-1 flex-col p-4">
          <MessageList messages={bubbles} isPending={processInformation.isPending} />

          {processInformation.isError ? (
            <div className="pb-3">
              <FormAlert variant="error">{getApiErrorMessage(processInformation.error)}</FormAlert>
            </div>
          ) : null}

          <Composer
            onSend={handleSend}
            disabled={processInformation.isPending || Boolean(savedInformationId)}
            placeholder={
              savedInformationId
                ? "保存が完了しました"
                : started
                  ? "続きを入力..."
                  : "共有したい情報を入力してください"
            }
          />
        </Card>

        {sources.length > 0 ? (
          <div className="flex flex-col gap-3">
            <div>
              <h2 className="mb-2 text-sm font-medium text-zinc-500">整理された項目</h2>
              <div className="flex flex-col gap-2">
                {sources.map((source) => (
                  <SourceCard key={source.key} source={source} />
                ))}
              </div>
            </div>

            <div>
              <p className="mb-1.5 text-xs text-zinc-500">
                {savedInformationId
                  ? "すべての項目が確定し、保存できます。"
                  : `確定した項目: ${sources.filter((s) => s.status === "confirmed").length} / ${sources.length}`}
              </p>
              <Button
                type="button"
                disabled={!savedInformationId}
                onClick={handleConfirmSave}
                className={savedInformationId ? "" : "blur-[1.5px]"}
              >
                {savedInformationId ? "保存して資料を確認する" : "すべて確定すると保存できます"}
              </Button>
            </div>
          </div>
        ) : null}
      </div>
    </div>
  );
}
