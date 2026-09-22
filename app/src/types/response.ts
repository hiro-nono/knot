// バックエンド usecase.ResponseItemView に対応する。
export interface ResponseItemView {
  source_id: string;
  option_id?: string;
  value?: string;
}

// バックエンド usecase.ResponseView に対応する。
// 匿名回答の場合user_idはundefinedになる。
export interface ResponseView {
  id: string;
  information_id: string;
  user_id?: string;
  items: ResponseItemView[];
  created_at: string;
  updated_at: string;
}

// POST /informations/:id/responses のリクエストボディ内の1件。
export interface SubmitResponseItemRequest {
  source_id: string;
  option_id?: string | null;
  value?: string | null;
}

// POST /informations/:id/responses のリクエストボディ。
export interface SubmitResponseRequest {
  items: SubmitResponseItemRequest[];
}
