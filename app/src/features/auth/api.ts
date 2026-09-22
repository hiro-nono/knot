import type { EmailOtpType } from "@supabase/supabase-js";

import { createClient } from "@/lib/auth/client";

// メール内の確認リンクの遷移先(app/auth/confirm/route.ts)を組み立てる。
// Supabaseダッシュボードのメールテンプレートが token_hash/type を含む形式で
// このURLを指すよう設定されている必要がある(README/セットアップ手順を参照)。
function buildEmailRedirectTo(next: string): string {
  const url = new URL("/auth/confirm", window.location.origin);
  url.searchParams.set("next", next);
  return url.toString();
}

export interface SignUpInput {
  email: string;
  password: string;
}

// サインアップ。成功するとSupabaseが確認メールを自動送信する
// (verify_email機能はその再送信のみを担当する)。
export async function signUp({ email, password }: SignUpInput) {
  const supabase = createClient();
  const { data, error } = await supabase.auth.signUp({
    email,
    password,
    options: { emailRedirectTo: buildEmailRedirectTo("/") },
  });
  if (error) {
    throw error;
  }
  return data;
}

export interface SignInInput {
  email: string;
  password: string;
}

// サインイン。成功するとSupabaseクライアントがaccess/refresh tokenを
// Cookieへ永続化し、以降の自動更新も行う。
export async function signIn({ email, password }: SignInInput) {
  const supabase = createClient();
  const { data, error } = await supabase.auth.signInWithPassword({ email, password });
  if (error) {
    throw error;
  }
  return data;
}

export interface RequestPasswordResetInput {
  email: string;
}

// パスワード再設定メールの送信依頼。
export async function requestPasswordReset({ email }: RequestPasswordResetInput) {
  const supabase = createClient();
  const { error } = await supabase.auth.resetPasswordForEmail(email, {
    redirectTo: buildEmailRedirectTo("/reset-password/confirm"),
  });
  if (error) {
    throw error;
  }
}

export interface ResetPasswordInput {
  password: string;
}

// パスワード再設定の確定。メール内リンク経由で確立された一時セッションを前提とする。
export async function resetPassword({ password }: ResetPasswordInput) {
  const supabase = createClient();
  const { data, error } = await supabase.auth.updateUser({ password });
  if (error) {
    throw error;
  }
  return data;
}

// パスワード再設定後、リンク経由の一時セッションを終了させるために使う。
export async function signOut() {
  const supabase = createClient();
  const { error } = await supabase.auth.signOut();
  if (error) {
    throw error;
  }
}

export interface ChangeEmailInput {
  email: string;
}

// メールアドレスの変更。認証済みユーザーのみ実行可能。
// 成功すると新しいメールアドレスに確認メールが送信される。
export async function changeEmail({ email }: ChangeEmailInput) {
  const supabase = createClient();
  const { data, error } = await supabase.auth.updateUser(
    { email },
    { emailRedirectTo: buildEmailRedirectTo("/") },
  );
  if (error) {
    throw error;
  }
  return data;
}

export interface ChangePasswordInput {
  password: string;
}

// パスワードの変更。認証済みユーザーのみ実行可能。
export async function changePassword({ password }: ChangePasswordInput) {
  const supabase = createClient();
  const { data, error } = await supabase.auth.updateUser({ password });
  if (error) {
    throw error;
  }
  return data;
}

export interface ResendVerificationEmailInput {
  // "signup": サインアップ後の確認メール、"email_change": メールアドレス変更後の確認メール。
  type: Extract<EmailOtpType, "signup" | "email_change">;
  email: string;
}

// 確認メールの再送信(signup/change_emailの成功後に使う)。
export async function resendVerificationEmail({ type, email }: ResendVerificationEmailInput) {
  const supabase = createClient();
  const { error } = await supabase.auth.resend({
    type,
    email,
    options: { emailRedirectTo: buildEmailRedirectTo("/") },
  });
  if (error) {
    throw error;
  }
}
