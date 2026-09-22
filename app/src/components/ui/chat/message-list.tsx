"use client";

import { useEffect, useRef } from "react";

import type { ChatBubble } from "./types";

export interface MessageListProps {
  messages: ChatBubble[];
  isPending?: boolean;
}

// Google/Gemini系のチャットに寄せた、シンプルな吹き出しリスト。
// 新しいメッセージが追加されたら自動で最下部へスクロールする。
export function MessageList({ messages, isPending = false }: MessageListProps) {
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth", block: "end" });
  }, [messages.length, isPending]);

  return (
    <div className="flex flex-1 flex-col gap-3 overflow-y-auto px-1 py-4">
      {messages.map((message) => (
        <div
          key={message.id}
          className={`flex ${message.role === "user" ? "justify-end" : "justify-start"}`}
        >
          <div
            className={`max-w-[85%] whitespace-pre-wrap rounded-2xl px-4 py-2.5 text-sm leading-relaxed ${
              message.role === "user"
                ? "bg-primary text-primary-foreground"
                : "bg-zinc-100 text-zinc-900"
            }`}
          >
            {message.content}
          </div>
        </div>
      ))}

      {isPending ? (
        <div className="flex justify-start">
          <div className="flex items-center gap-1 rounded-2xl bg-zinc-100 px-4 py-3">
            <span className="h-1.5 w-1.5 animate-bounce rounded-full bg-zinc-400 [animation-delay:-0.3s]" />
            <span className="h-1.5 w-1.5 animate-bounce rounded-full bg-zinc-400 [animation-delay:-0.15s]" />
            <span className="h-1.5 w-1.5 animate-bounce rounded-full bg-zinc-400" />
          </div>
        </div>
      ) : null}

      <div ref={bottomRef} />
    </div>
  );
}
