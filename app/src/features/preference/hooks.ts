import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { applyPreference, chat, generateComparison, generateDisplay } from "@/features/preference/api";
import { preferenceKeys } from "@/features/preference/query-keys";
import type { ApplyPreferenceRequest, GenerateDisplayResponse } from "@/types/display";

// POST /informations/:id/display
// 副作用のない生成なので、react-queryのキャッシュ(GET相当)として扱う。
export function useDisplay(informationId: string, enabled = true) {
  return useQuery({
    queryKey: preferenceKeys.display(informationId),
    queryFn: () => generateDisplay(informationId),
    enabled,
  });
}

// POST /informations/:id/display/comparisons
export function useGenerateComparison(informationId: string) {
  return useMutation({
    mutationFn: (body: Parameters<typeof generateComparison>[1]) =>
      generateComparison(informationId, body),
  });
}

// POST /informations/:id/display/comparisons/apply の呼び出しに加え、
// 採用したパターンの表示内容(既に比較生成で取得済み)を渡してもらう。
interface ApplyPreferenceVariables extends ApplyPreferenceRequest {
  appliedDisplay: GenerateDisplayResponse;
}

// POST /informations/:id/display/comparisons/apply
//
// 選んだパターンの内容は呼び出し元(コンポーネント)が既に持っているため、
// 採用時にGenerateDisplayを再度呼ばず(=AI呼び出しを増やさず)、
// そのままキャッシュへ書き込んで表示に反映する。
export function useApplyPreference(informationId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (variables: ApplyPreferenceVariables) =>
      applyPreference(informationId, { key: variables.key, value: variables.value }),
    onSuccess: (_data, variables) => {
      queryClient.setQueryData(preferenceKeys.display(informationId), variables.appliedDisplay);
    },
  });
}

// POST /informations/:id/display/chat
// 資料の情報についてのQ&Aであり、Preferenceは更新されないため
// 表示のキャッシュを無効化する必要はない。
export function useChat(informationId: string) {
  return useMutation({
    mutationFn: (body: Parameters<typeof chat>[1]) => chat(informationId, body),
  });
}
