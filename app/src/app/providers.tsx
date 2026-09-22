"use client";

import { QueryClientProvider } from "@tanstack/react-query";
import { useEffect, useState } from "react";

import { getCsrfToken } from "@/lib/api/csrf";
import { makeQueryClient } from "@/lib/api/query-client";

export function Providers({ children }: { children: React.ReactNode }) {
  // useStateで初期化することで、Server Componentのレンダー間で
  // QueryClientが共有されるのを防ぐ(Next.js App Routerでの既知の注意点)。
  const [queryClient] = useState(() => makeQueryClient());

  useEffect(() => {
    // 初回のミューテーションでCSRFトークン発行の往復が発生しないよう先読みする。
    // 失敗しても各リクエスト側でオンデマンドに取得し直すため無視してよい。
    getCsrfToken().catch(() => {});
  }, []);

  return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
}
