import { env } from "@/lib/env";

// バックエンド internal/infrastructure/http/csrf と対応する
// 二重提出Cookie(double-submit cookie)方式のCSRF対策。
// GET /csrf-token がcsrf_token Cookie(HttpOnly=false)を発行し、
// クライアントはPOST/PATCH/DELETEのたびに同じ値をX-CSRF-Tokenヘッダーで返す。
const COOKIE_NAME = "csrf_token";
export const CSRF_HEADER_NAME = "X-CSRF-Token";

let cachedToken: string | undefined;
let inFlight: Promise<string> | undefined;

function readCookie(name: string): string | undefined {
  if (typeof document === "undefined") {
    return undefined;
  }
  const match = document.cookie.match(new RegExp(`(?:^|; )${name}=([^;]*)`));
  return match ? decodeURIComponent(match[1]) : undefined;
}

async function fetchCsrfToken(): Promise<string> {
  const base = env.apiBaseUrl().replace(/\/+$/, "");
  const response = await fetch(`${base}/csrf-token`, {
    method: "GET",
    credentials: "include",
  });

  if (!response.ok) {
    throw new Error(`failed to issue csrf token: ${response.status}`);
  }

  const data = (await response.json()) as { csrf_token: string };
  cachedToken = data.csrf_token;
  return cachedToken;
}

// 現在のCSRFトークンを返す。ブラウザ以外(Server Component/Action)からは
// Cookieベースの二重提出方式が意味を持たないため常にundefinedを返す。
// forceRefresh=trueの場合、Cookieの値を信用せずサーバーから新規発行を受ける
// (バックエンドがトークン不一致で403を返した後のリトライに使う)。
export async function getCsrfToken(forceRefresh = false): Promise<string | undefined> {
  if (typeof window === "undefined") {
    return undefined;
  }

  if (!forceRefresh) {
    const fromCookie = readCookie(COOKIE_NAME);
    if (fromCookie) {
      cachedToken = fromCookie;
      return cachedToken;
    }
    if (cachedToken) {
      return cachedToken;
    }
  } else {
    cachedToken = undefined;
  }

  if (!inFlight) {
    inFlight = fetchCsrfToken().finally(() => {
      inFlight = undefined;
    });
  }

  return inFlight;
}
