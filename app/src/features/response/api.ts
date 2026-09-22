import { apiClient } from "@/lib/api/client";
import type { ResponseView, SubmitResponseRequest } from "@/types/response";

// POST /informations/:id/responses
// PUBLIC+ANONYMOUSなInformationは未認証でも呼び出せる
// (apiClientはSupabaseセッションがあれば自動でBearerトークンを付与し、
// なければ付与せずに送信する)。
export function submitResponse(informationId: string, body: SubmitResponseRequest) {
  return apiClient.post<ResponseView>(`/informations/${informationId}/responses`, body);
}

// GET /informations/:id/responses (そのInformationを所有するAccountのみ)
export function listResponses(informationId: string) {
  return apiClient.get<ResponseView[]>(`/informations/${informationId}/responses`);
}
