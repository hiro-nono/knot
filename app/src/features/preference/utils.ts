import type { ChatResponse } from "@/types/display";

// display/chatも同様に、assistant側の生のmessages内容は人間向けではないため
// 使わない。confirmation_question(継続適用してよいか確認する質問)があれば
// それを、無ければpreference_updatedの結果に応じた短い通知文を返す。
export function describeAssistantTurn(response: ChatResponse): string {
  if (response.needs_confirmation && response.confirmation_question) {
    return response.confirmation_question;
  }

  return response.preference_updated
    ? "表示の好みを保存しました。今後も同じ調整を適用します。"
    : "表示を一時的に調整しました。";
}
