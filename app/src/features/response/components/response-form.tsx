"use client";

import { type FormEvent, useState } from "react";

import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { FormAlert } from "@/components/ui/form-alert";
import { useSources } from "@/features/information/hooks";
import { useSubmitResponse } from "@/features/response/hooks";
import { getApiErrorMessage } from "@/lib/api/errors";
import type { SourceView } from "@/types/information";
import type { SubmitResponseItemRequest } from "@/types/response";

type Answer = string | string[];

function buildItems(sources: SourceView[], answers: Record<string, Answer>): SubmitResponseItemRequest[] {
  const items: SubmitResponseItemRequest[] = [];

  for (const source of sources) {
    const answer = answers[source.id];

    if (source.interaction_type === "text") {
      if (typeof answer === "string" && answer.trim()) {
        items.push({ source_id: source.id, value: answer.trim() });
      }
    } else if (source.interaction_type === "radio") {
      if (typeof answer === "string" && answer) {
        items.push({ source_id: source.id, option_id: answer });
      }
    } else if (source.interaction_type === "check" && Array.isArray(answer)) {
      for (const optionId of answer) {
        items.push({ source_id: source.id, option_id: optionId });
      }
    }
  }

  return items;
}

const INPUT_CLASS =
  "w-full rounded-md border border-zinc-300 px-3 py-2 text-sm text-zinc-900 outline-none focus:ring-2 focus:ring-primary";

// GET /informations/:id/sources で取得したinteraction_type付きのSourceから
// 回答フォームを組み立て、POST /informations/:id/responses で送信する。
export function ResponseForm({ informationId }: { informationId: string }) {
  const sourcesQuery = useSources(informationId);
  const submitResponse = useSubmitResponse(informationId);
  const [answers, setAnswers] = useState<Record<string, Answer>>({});

  const answerable = (sourcesQuery.data ?? []).filter((source) => source.interaction_type);

  if (sourcesQuery.isLoading || answerable.length === 0) {
    return null;
  }

  function setRadio(sourceId: string, optionId: string) {
    setAnswers((prev) => ({ ...prev, [sourceId]: optionId }));
  }

  function toggleCheck(sourceId: string, optionId: string, checked: boolean) {
    setAnswers((prev) => {
      const current = Array.isArray(prev[sourceId]) ? (prev[sourceId] as string[]) : [];
      const next = checked ? [...current, optionId] : current.filter((id) => id !== optionId);
      return { ...prev, [sourceId]: next };
    });
  }

  function setText(sourceId: string, value: string) {
    setAnswers((prev) => ({ ...prev, [sourceId]: value }));
  }

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const items = buildItems(answerable, answers);
    if (items.length === 0) {
      return;
    }
    submitResponse.mutate({ items });
  }

  if (submitResponse.isSuccess) {
    return (
      <div>
        <h2 className="mb-2 text-sm font-medium text-zinc-500">回答する</h2>
        <Card className="p-4">
          <FormAlert variant="success">回答を送信しました。ありがとうございました。</FormAlert>
        </Card>
      </div>
    );
  }

  return (
    <div>
      <h2 className="mb-2 text-sm font-medium text-zinc-500">回答する</h2>
      <Card className="flex flex-col gap-4 p-4">
        <form onSubmit={handleSubmit} className="flex flex-col gap-5">
          {answerable.map((source) => (
            <div key={source.id} className="flex flex-col gap-2">
              <p className="text-sm font-medium text-zinc-900">{source.value || source.key}</p>

              {source.interaction_type === "text" ? (
                <input
                  type="text"
                  className={INPUT_CLASS}
                  value={(answers[source.id] as string) ?? ""}
                  onChange={(e) => setText(source.id, e.target.value)}
                />
              ) : null}

              {source.interaction_type === "radio"
                ? source.options?.map((option) => (
                    <label key={option.id} className="flex items-center gap-2 text-sm text-zinc-700">
                      <input
                        type="radio"
                        name={source.id}
                        value={option.id}
                        checked={answers[source.id] === option.id}
                        onChange={() => setRadio(source.id, option.id)}
                        className="accent-primary"
                      />
                      {option.value}
                    </label>
                  ))
                : null}

              {source.interaction_type === "check"
                ? source.options?.map((option) => (
                    <label key={option.id} className="flex items-center gap-2 text-sm text-zinc-700">
                      <input
                        type="checkbox"
                        value={option.id}
                        checked={
                          Array.isArray(answers[source.id]) &&
                          (answers[source.id] as string[]).includes(option.id)
                        }
                        onChange={(e) => toggleCheck(source.id, option.id, e.target.checked)}
                        className="accent-primary"
                      />
                      {option.value}
                    </label>
                  ))
                : null}
            </div>
          ))}

          {submitResponse.isError ? (
            <FormAlert variant="error">{getApiErrorMessage(submitResponse.error)}</FormAlert>
          ) : null}

          <Button type="submit" isLoading={submitResponse.isPending} className="w-auto self-start px-6">
            回答を送信
          </Button>
        </form>
      </Card>
    </div>
  );
}
