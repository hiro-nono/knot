import type { Message } from "@/types/message";

// バックエンド domain.InformationAccessType に対応する。
export type InformationAccessType = "public" | "restricted";

// バックエンド domain.InformationResponsePolicy に対応する。
export type InformationResponsePolicy = "anonymous" | "authenticated";

// バックエンド usecase.StructuredOption に対応する。
export interface StructuredOption {
  value: string;
  sort_order: number;
}

// バックエンド usecase.StructuredSource に対応する。
export interface StructuredSource {
  type: string;
  key: string;
  value: string;
  status: string;
  question?: string;
  interaction_type?: string;
  options?: StructuredOption[];
}

// バックエンド usecase.StructuredInformation に対応する。
export interface StructuredInformation {
  title: string;
  sources: StructuredSource[];
}

// POST /informations/messages のリクエストボディ。
export interface ProcessInformationRequest {
  access_type: InformationAccessType;
  response_policy: InformationResponsePolicy;
  messages: Message[];
  user_input: string;
}

// POST /informations/messages のレスポンス。
// confirmedがtrueの場合のみinformation_idが設定される。
export interface ProcessInformationResponse {
  messages: Message[];
  structured: StructuredInformation | null;
  confirmed: boolean;
  information_id?: string;
}

// GET /informations/:id/recipients のレスポンス。
export interface ListRecipientsResponse {
  user_ids: string[];
}

// バックエンド usecase.OptionView に対応する。
export interface OptionView {
  id: string;
  value: string;
  sort_order: number;
}

// バックエンド usecase.SourceView に対応する。
// GET /informations/:id/sources のレスポンス項目。
// SubmitResponse(POST /informations/:id/responses)に必要なsource_id・
// interaction_type・option_idを受信者に提示するために使う。
export interface SourceView {
  id: string;
  type: string;
  key: string;
  value: string;
  interaction_type?: "radio" | "check" | "text";
  options?: OptionView[];
}

// バックエンド usecase.InformationSummaryView に対応する。
// GET /informations のレスポンス項目。
export interface InformationSummaryView {
  id: string;
  title: string;
  access_type: InformationAccessType;
  response_policy: InformationResponsePolicy;
  created_at: string;
}
