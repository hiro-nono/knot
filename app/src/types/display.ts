import type { Message } from "@/types/message";

// バックエンド usecase.DisplayContent に対応する。
export interface DisplayContent {
  title: string;
  body: string;
}

// バックエンド usecase.GenerateDisplayOutput に対応する。
// POST /informations/:id/display のレスポンス。
// has_preferenceがfalseの場合、受信者はまだ何もPreferenceを記録していない
// ということなので、クライアント側はA/B比較を提示するきっかけに使う。
export interface GenerateDisplayResponse extends DisplayContent {
  has_preference: boolean;
}

// バックエンド usecase.ComparisonPattern に対応する。
// valueは受信者に見せる表示ではなく、このパターンが採用した傾向を表す
// メタデータ(選択結果をPreferenceとして記録するためのタグ)。
export interface ComparisonPattern {
  value: string;
  display: DisplayContent;
}

// POST /informations/:id/display/comparisons のリクエストボディ。
// 比較する項目(key)だけを指定する。どんな値(傾向)で対照させるかは
// アプリ(AI)側が目的を持って決める。
export interface GenerateComparisonRequest {
  key: string;
}

// POST /informations/:id/display/comparisons のレスポンス。
export interface GenerateComparisonResponse {
  key: string;
  pattern_a: ComparisonPattern;
  pattern_b: ComparisonPattern;
}

// POST /informations/:id/display/comparisons/apply のリクエストボディ。
// 選んだパターンのvalueをPreferenceとして採用する(AI呼び出しは行わない)。
export interface ApplyPreferenceRequest {
  key: string;
  value: string;
}

// POST /informations/:id/display/chat のリクエストボディ。
// この機能は資料の情報についての質問(Q&A)であり、表示の見せ方を調整する
// ものではない。表示の好み(Preference)はA/B比較の選択結果からのみ更新される。
export interface ChatRequest {
  messages: Message[];
  user_input: string;
}

// POST /informations/:id/display/chat のレスポンス。
export interface ChatResponse {
  messages: Message[];
  answer: string;
}
