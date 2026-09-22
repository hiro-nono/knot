"use client";

import { type KeyboardEvent, useState } from "react";

export interface ComposerProps {
  onSend: (text: string) => void;
  disabled?: boolean;
  placeholder?: string;
}

// Google検索バーのような、角丸のピル型入力欄。
// Enterで送信、Shift+Enterで改行。
export function Composer({ onSend, disabled = false, placeholder }: ComposerProps) {
  const [value, setValue] = useState("");

  function submit() {
    const text = value.trim();
    if (!text || disabled) {
      return;
    }
    onSend(text);
    setValue("");
  }

  function handleKeyDown(event: KeyboardEvent<HTMLTextAreaElement>) {
    if (event.key === "Enter" && !event.shiftKey) {
      event.preventDefault();
      submit();
    }
  }

  return (
    <div className="flex items-end gap-2 rounded-3xl border border-zinc-200 bg-white px-4 py-2 shadow-sm transition-shadow focus-within:shadow-md">
      <textarea
        rows={1}
        value={value}
        onChange={(e) => setValue(e.target.value)}
        onKeyDown={handleKeyDown}
        disabled={disabled}
        placeholder={placeholder}
        className="max-h-40 flex-1 resize-none bg-transparent py-1.5 text-sm text-zinc-900 outline-none placeholder:text-zinc-400 disabled:opacity-50"
      />
      <button
        type="button"
        onClick={submit}
        disabled={disabled || !value.trim()}
        aria-label="送信"
        className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-primary text-primary-foreground transition-colors hover:bg-primary-hover disabled:cursor-not-allowed disabled:bg-zinc-300"
      >
        <svg viewBox="0 0 24 24" fill="none" className="h-4 w-4" aria-hidden="true">
          <path
            d="M3.4 20.6 21 12 3.4 3.4 3 10l12 2-12 2z"
            fill="currentColor"
          />
        </svg>
      </button>
    </div>
  );
}
