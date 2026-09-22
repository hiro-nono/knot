"use client";

import type { Session } from "@supabase/supabase-js";
import { useEffect, useState } from "react";

import { createClient } from "@/lib/auth/client";

export interface SupabaseSessionState {
  session: Session | null;
  isLoading: boolean;
}

// Client Componentから現在のSupabaseセッションを参照するためのhook。
// onAuthStateChangeを購読し、サインイン・サインアウト・トークン更新に追従する。
export function useSupabaseSession(): SupabaseSessionState {
  const [session, setSession] = useState<Session | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const supabase = createClient();

    supabase.auth.getSession().then(({ data }) => {
      setSession(data.session);
      setIsLoading(false);
    });

    const {
      data: { subscription },
    } = supabase.auth.onAuthStateChange((_event, nextSession) => {
      setSession(nextSession);
      setIsLoading(false);
    });

    return () => subscription.unsubscribe();
  }, []);

  return { session, isLoading };
}
