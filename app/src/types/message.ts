// AIとの対話1ターン(発言者と本文)を表す。
// バックエンド usecase.Message に対応する。
export interface Message {
  role: string;
  content: string;
}
