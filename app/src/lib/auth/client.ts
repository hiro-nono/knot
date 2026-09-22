import { createBrowserClient } from "@supabase/ssr";

import { env } from "@/lib/env";

// ブラウザ(Client Component)から使うSupabaseクライアント。
// セッションはCookieに保存され、middleware.tsが更新を担う。
export function createClient() {
  return createBrowserClient(env.supabaseUrl(), env.supabaseAnonKey());
}
