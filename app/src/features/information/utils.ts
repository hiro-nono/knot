import type { ProcessInformationResponse } from "@/types/information";

// バックエンドが返すmessages(usecase.Message[])は、assistant側の内容が
// 生のJSON(構造化結果そのもの)であり人間向けの文面ではないため、チャット表示には使わない。
// 代わりにstructured.sourcesのquestion(未確定項目についてAIが尋ねる内容)から
// 人間向けの返信を組み立てる。
export function describeAssistantTurn(response: ProcessInformationResponse): string {
  if (response.confirmed) {
    return `「${response.structured?.title ?? ""}」の内容を確定しました。`;
  }

  const pendingQuestions =
    response.structured?.sources
      .filter((source) => source.status !== "confirmed" && source.question)
      .map((source) => source.question as string) ?? [];

  if (pendingQuestions.length > 0) {
    return pendingQuestions.join("\n");
  }

  return "内容を確認しています。もう少し詳しく教えてください。";
}
