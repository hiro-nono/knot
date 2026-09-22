// Supabase Auth(GoTrue)が返す代表的な英語メッセージを日本語に変換する。
// 未知のメッセージはそのまま表示する(fail-open。エラーを握りつぶさない)。
const KNOWN_MESSAGES: Record<string, string> = {
  "Invalid login credentials": "メールアドレスまたはパスワードが正しくありません",
  "Email not confirmed": "メールアドレスが確認されていません。確認メールをご確認ください",
  "User already registered": "このメールアドレスは既に登録されています",
  "New password should be different from the old password.":
    "新しいパスワードは現在のパスワードと異なるものにしてください",
};

const RATE_LIMIT_PATTERN = /^For security purposes, you can only request this after (\d+) seconds\.$/;

export function getAuthErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    const rateLimitMatch = error.message.match(RATE_LIMIT_PATTERN);
    if (rateLimitMatch) {
      return `しばらく時間をおいてから再度お試しください(あと${rateLimitMatch[1]}秒)`;
    }
    return KNOWN_MESSAGES[error.message] ?? error.message;
  }
  return "予期しないエラーが発生しました";
}
