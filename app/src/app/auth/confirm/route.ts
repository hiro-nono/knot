import type { EmailOtpType } from "@supabase/supabase-js";
import { type NextRequest, NextResponse } from "next/server";

import { createClient } from "@/lib/auth/server";

// Supabaseの確認メール(signup/email_change/recovery)内リンクの遷移先。
//
// Supabaseダッシュボード(Authentication > Email Templates)側で、各テンプレートの
// リンクを以下の形式に変更しておく必要がある(デフォルトの{{ .ConfirmationURL }}のままでは
// このRoute Handlerを経由しない):
//   {{ .SiteURL }}/auth/confirm?token_hash={{ .TokenHash }}&type={{ .Type }}&next=...
//
// token_hash/typeが無い場合(PKCEのcode交換を使う構成向け)はcodeパラメータも
// フォールバックとして扱う。
export async function GET(request: NextRequest) {
  const { searchParams, origin } = new URL(request.url);
  const tokenHash = searchParams.get("token_hash");
  const type = searchParams.get("type") as EmailOtpType | null;
  const code = searchParams.get("code");
  const next = searchParams.get("next") ?? "/";

  const supabase = await createClient();

  if (tokenHash && type) {
    const { error } = await supabase.auth.verifyOtp({ type, token_hash: tokenHash });
    if (!error) {
      return NextResponse.redirect(new URL(redirectPathFor(type, next), origin));
    }
  } else if (code) {
    const { error } = await supabase.auth.exchangeCodeForSession(code);
    if (!error) {
      return NextResponse.redirect(new URL(next, origin));
    }
  }

  return NextResponse.redirect(new URL("/signin?error=confirm_failed", origin));
}

// recoveryはパスワード再設定フォームへ、それ以外(signup/email_change)は
// 呼び出し元が指定したnext(未指定ならホーム)へ遷移する。
function redirectPathFor(type: EmailOtpType, next: string): string {
  if (type === "recovery") {
    return "/reset-password/confirm";
  }
  return next;
}
