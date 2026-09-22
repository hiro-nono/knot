import { apiClient } from "@/lib/api/client";
import type {
  InformationSummaryView,
  ListRecipientsResponse,
  ProcessInformationRequest,
  ProcessInformationResponse,
  SourceView,
} from "@/types/information";

// GET /informations
// 自分(発信者)が過去に作成したInformationを一覧取得する。
export function listMyInformations() {
  return apiClient.get<InformationSummaryView[]>("/informations");
}

// POST /informations/messages
// 発信者とAIの対話を1ターン進める。
export function processInformation(body: ProcessInformationRequest) {
  return apiClient.post<ProcessInformationResponse>("/informations/messages", body);
}

// GET /informations/:id/sources
// 受信者がResponseを送信する際に必要なsource_id・interaction_type・option_idを取得する。
export function listSources(informationId: string) {
  return apiClient.get<SourceView[]>(`/informations/${informationId}/sources`);
}

// POST /informations/:id/recipients
export function addRecipient(informationId: string, userId: string) {
  return apiClient.post<void>(`/informations/${informationId}/recipients`, { user_id: userId });
}

// GET /informations/:id/recipients
export function listRecipients(informationId: string) {
  return apiClient.get<ListRecipientsResponse>(`/informations/${informationId}/recipients`);
}

// DELETE /informations/:id/recipients/:user_id
export function removeRecipient(informationId: string, userId: string) {
  return apiClient.delete<void>(`/informations/${informationId}/recipients/${userId}`);
}
