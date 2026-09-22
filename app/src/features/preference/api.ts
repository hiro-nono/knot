import { apiClient } from "@/lib/api/client";
import type {
  ChatRequest,
  ChatResponse,
  DisplayContent,
  PrepareComparisonRequest,
  PrepareComparisonResponse,
  SelectComparisonRequest,
} from "@/types/display";

// POST /informations/:id/display
// 受信者(認証済みUser)のPreferenceに応じて最適化した表示を生成する。
export function generateDisplay(informationId: string) {
  return apiClient.post<DisplayContent>(`/informations/${informationId}/display`);
}

// POST /informations/:id/display/comparisons
// 指定したPreferenceキーの値をA/Bそれぞれに変えた表示を2パターン生成する。
export function prepareComparison(informationId: string, body: PrepareComparisonRequest) {
  return apiClient.post<PrepareComparisonResponse>(
    `/informations/${informationId}/display/comparisons`,
    body,
  );
}

// POST /informations/:id/display/comparisons/selection
// 受信者が選んだA/Bの結果をもとにPreferenceを更新する。
export function selectComparison(informationId: string, body: SelectComparisonRequest) {
  return apiClient.post<void>(`/informations/${informationId}/display/comparisons/selection`, body);
}

// POST /informations/:id/display/chat
// 受信者からの表示調整の要望を1ターン処理する。
export function chat(informationId: string, body: ChatRequest) {
  return apiClient.post<ChatResponse>(`/informations/${informationId}/display/chat`, body);
}
