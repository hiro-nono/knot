import { createServerClient } from "@supabase/ssr";
import { type NextRequest, NextResponse } from "next/server";

import { env } from "@/lib/env";

// 認証なしでアクセスできるパス。ここに無いパスは、Supabaseセッションが
// 無ければ /signin へリダイレクトする(多層防御。実データの取得・変更は
// 最終的にバックエンドのauthMiddlewareが必ず守るため、ここはクライアント側の
// ガード(DashboardLayout等)が効かない場合の保険)。
//
// change-password・change-email・reset-password/confirmは、ここでは保護せず
// 各ページ自身のuseSupabaseSessionチェックに委ねる(リンク切れ・要ログインの
// 専用メッセージを表示するため、汎用リダイレクトで上書きしない)。
// onboardingは意図的に含めない(サインイン直後の一時的な導線であり、
// 必ずセッションを持っているはずなので保護してよい)。
const PUBLIC_PATHS = new Set([
  "/signin",
  "/signup",
  "/reset-password",
  "/reset-password/confirm",
  "/verify-email",
  "/change-password",
  "/change-email",
]);

function isPublicPath(pathname: string): boolean {
  if (PUBLIC_PATHS.has(pathname)) {
    return true;
  }
  if (pathname.startsWith("/auth/")) {
    return true;
  }
  // PUBLIC+ANONYMOUSなInformationへの回答ページ(認証不要)。
  if (/^\/informations\/[^/]+\/respond$/.test(pathname)) {
    return true;
  }
  return false;
}

// Next.jsのmiddleware.tsから呼び出し、Supabaseセッション(Cookie)を
// リクエストごとに検証・更新する。Server Componentは Cookie を書き換えられないため、
// この処理がセッションの延命を担う。
//
// あわせて、保護対象パスへの未認証リクエストを/signinへリダイレクトする。
export async function updateSession(request: NextRequest) {
  let response = NextResponse.next({ request });

  const supabase = createServerClient(env.supabaseUrl(), env.supabaseAnonKey(), {
    cookies: {
      getAll() {
        return request.cookies.getAll();
      },
      setAll(cookiesToSet) {
        for (const { name, value } of cookiesToSet) {
          request.cookies.set(name, value);
        }
        response = NextResponse.next({ request });
        for (const { name, value, options } of cookiesToSet) {
          response.cookies.set(name, value, options);
        }
      },
    },
  });

  // getUser()はSupabase Auth serverに問い合わせて検証するため、
  // getSession()より安全(JWTをローカルでデコードするだけの検証を避ける)。
  const {
    data: { user },
  } = await supabase.auth.getUser();

  if (!user && !isPublicPath(request.nextUrl.pathname)) {
    const redirectUrl = request.nextUrl.clone();
    redirectUrl.pathname = "/signin";
    redirectUrl.searchParams.set("next", request.nextUrl.pathname);
    return NextResponse.redirect(redirectUrl);
  }

  return response;
}
