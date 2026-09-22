import { env } from "@/lib/env";
import { createClient as createBrowserSupabaseClient } from "@/lib/auth/client";

import { CSRF_HEADER_NAME, getCsrfToken } from "./csrf";
import { ApiError } from "./errors";

const MUTATING_METHODS = new Set(["POST", "PATCH", "DELETE"]);

export interface ApiRequestOptions {
  method?: "GET" | "POST" | "PATCH" | "DELETE";
  body?: unknown;
  searchParams?: Record<string, string | undefined>;
  // Server Component/Action から呼ぶ場合、ブラウザのセッションは参照できないため
  // 呼び出し元がSupabaseサーバークライアントから取得したaccess_tokenを渡す。
  accessToken?: string;
  signal?: AbortSignal;
}

function buildUrl(path: string, searchParams?: Record<string, string | undefined>): string {
  const base = env.apiBaseUrl().replace(/\/+$/, "");
  const url = new URL(base + path);

  if (searchParams) {
    for (const [key, value] of Object.entries(searchParams)) {
      if (value !== undefined) {
        url.searchParams.set(key, value);
      }
    }
  }

  return url.toString();
}

// ブラウザ実行時のみ、現在のSupabaseセッションからaccess_tokenを取得する。
// Server Component/Actionから呼ばれた場合はundefinedを返す
// (代わりにApiRequestOptions.accessTokenを明示的に渡すこと)。
async function getBrowserAccessToken(): Promise<string | undefined> {
  if (typeof window === "undefined") {
    return undefined;
  }

  const supabase = createBrowserSupabaseClient();
  const {
    data: { session },
  } = await supabase.auth.getSession();

  return session?.access_token;
}

function isErrorBody(value: unknown): value is { error: string } {
  return (
    typeof value === "object" &&
    value !== null &&
    "error" in value &&
    typeof (value as { error: unknown }).error === "string"
  );
}

// バックエンドのcsrfミドルウェアが返す403(トークン不一致・欠落)かどうかを判定する。
// domain.ErrForbidden由来の403(権限不足)と区別し、CSRFの場合のみ
// トークンを取り直して1回だけリトライする。
function isCsrfRejection(status: number, data: unknown): boolean {
  if (status !== 403 || !isErrorBody(data)) {
    return false;
  }
  return data.error === "missing csrf cookie" || data.error === "invalid csrf token";
}

// バックエンド(knot-api、gin)への共通フェッチャー。
// 認証済みエンドポイントには、認証済みユーザーのSupabase access_tokenを
// Authorization: Bearer ヘッダーとして必ず付与する。
// 状態変更リクエスト(POST/PATCH/DELETE)にはCSRFトークン(X-CSRF-Token)も
// 付与し、CORSでCookie(csrf_token)を送受信できるようcredentials: "include"を使う。
export async function apiFetch<T>(
  path: string,
  options: ApiRequestOptions = {},
  isRetry = false,
): Promise<T> {
  const { method = "GET", body, searchParams, signal } = options;
  const accessToken = options.accessToken ?? (await getBrowserAccessToken());

  const headers: Record<string, string> = { Accept: "application/json" };
  if (body !== undefined) {
    headers["Content-Type"] = "application/json";
  }
  if (accessToken) {
    headers.Authorization = `Bearer ${accessToken}`;
  }
  if (MUTATING_METHODS.has(method)) {
    const csrfToken = await getCsrfToken();
    if (csrfToken) {
      headers[CSRF_HEADER_NAME] = csrfToken;
    }
  }

  const response = await fetch(buildUrl(path, searchParams), {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
    signal,
    credentials: "include",
  });

  if (response.status === 204) {
    return undefined as T;
  }

  const text = await response.text();
  const data: unknown = text ? JSON.parse(text) : undefined;

  if (!response.ok) {
    if (!isRetry && isCsrfRejection(response.status, data)) {
      await getCsrfToken(true);
      return apiFetch<T>(path, options, true);
    }

    const message = isErrorBody(data) ? data.error : response.statusText;
    throw new ApiError(message, response.status);
  }

  return data as T;
}

export const apiClient = {
  get: <T>(path: string, options?: Omit<ApiRequestOptions, "method" | "body">) =>
    apiFetch<T>(path, { ...options, method: "GET" }),
  post: <T>(path: string, body?: unknown, options?: Omit<ApiRequestOptions, "method" | "body">) =>
    apiFetch<T>(path, { ...options, method: "POST", body }),
  patch: <T>(path: string, body?: unknown, options?: Omit<ApiRequestOptions, "method" | "body">) =>
    apiFetch<T>(path, { ...options, method: "PATCH", body }),
  delete: <T>(path: string, options?: Omit<ApiRequestOptions, "method" | "body">) =>
    apiFetch<T>(path, { ...options, method: "DELETE" }),
};
