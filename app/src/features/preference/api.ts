import { apiClient } from "@/lib/api/client";
import type {
  ApplyPreferenceRequest,
  ChatRequest,
  ChatResponse,
  GenerateComparisonRequest,
  GenerateComparisonResponse,
  GenerateDisplayResponse,
} from "@/types/display";

// POST /informations/:id/display
// 受信者(認証済みUser)のPreferenceに応じて最適化した表示を生成する。
export function generateDisplay(informationId: string) {
  return apiClient.post<GenerateDisplayResponse>(`/informations/${informationId}/display`);
}

// POST /informations/:id/display/comparisons
// 指定したPreferenceキーについて、AIが決めた対照的な2パターン(A/B)の表示を生成する。
export function generateComparison(informationId: string, body: GenerateComparisonRequest) {
  return apiClient.post<GenerateComparisonResponse>(
    `/informations/${informationId}/display/comparisons`,
    body,
  );
}

// POST /informations/:id/display/comparisons/apply
// 選んだパターンの値をPreferenceとして採用する。
export function applyPreference(informationId: string, body: ApplyPreferenceRequest) {
  return apiClient.post<void>(`/informations/${informationId}/display/comparisons/apply`, body);
}

// POST /informations/:id/display/chat
// 受信者からの、資料の情報についての質問に1ターン回答する
// (表示の見せ方を調整する機能ではない)。
export function chat(informationId: string, body: ChatRequest) {
  return apiClient.post<ChatResponse>(`/informations/${informationId}/display/chat`, body);
}
