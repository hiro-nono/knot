import type { Message } from "@/types/message";

// バックエンド usecase.DisplayContent に対応する。
export interface DisplayContent {
  title: string;
  body: string;
}

// バックエンド usecase.PreferenceComparison に対応する。
export interface PreferenceComparison {
  key: string;
  value_a: string;
  value_b: string;
}

// POST /informations/:id/display/comparisons のリクエストボディ。
export type PrepareComparisonRequest = PreferenceComparison;

// POST /informations/:id/display/comparisons のレスポンス。
export interface PrepareComparisonResponse {
  display_a: DisplayContent;
  display_b: DisplayContent;
  comparison: PreferenceComparison;
}

// POST /informations/:id/display/comparisons/selection のリクエストボディ。
export interface SelectComparisonRequest {
  comparison: PreferenceComparison;
  selected: "a" | "b";
}

// POST /informations/:id/display/chat のリクエストボディ。
export interface ChatRequest {
  messages: Message[];
  user_input: string;
}

// POST /informations/:id/display/chat のレスポンス。
export interface ChatResponse {
  messages: Message[];
  display: DisplayContent;
  needs_confirmation: boolean;
  confirmation_question?: string;
  preference_updated: boolean;
}
