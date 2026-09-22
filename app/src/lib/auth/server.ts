import { createServerClient } from "@supabase/ssr";
import { cookies } from "next/headers";

import { env } from "@/lib/env";

// Server Component / Route Handler / Server Action から使うSupabaseクライアント。
// Server Componentからの呼び出しではCookieの書き込みができないため、
// setAllは失敗を無視する(セッション更新はmiddleware.tsが担う)。
export async function createClient() {
  const cookieStore = await cookies();

  return createServerClient(env.supabaseUrl(), env.supabaseAnonKey(), {
    cookies: {
      getAll() {
        return cookieStore.getAll();
      },
      setAll(cookiesToSet) {
        try {
          for (const { name, value, options } of cookiesToSet) {
            cookieStore.set(name, value, options);
          }
        } catch {
          // Server Componentから呼ばれた場合はここに来るが、
          // middlewareがセッション更新を行うため無視してよい。
        }
      },
    },
  });
}
