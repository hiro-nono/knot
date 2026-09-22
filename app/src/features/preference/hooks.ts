import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { chat, generateDisplay, prepareComparison, selectComparison } from "@/features/preference/api";
import { preferenceKeys } from "@/features/preference/query-keys";

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
export function usePrepareComparison(informationId: string) {
  return useMutation({
    mutationFn: (body: Parameters<typeof prepareComparison>[1]) =>
      prepareComparison(informationId, body),
  });
}

// POST /informations/:id/display/comparisons/selection
export function useSelectComparison(informationId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (body: Parameters<typeof selectComparison>[1]) =>
      selectComparison(informationId, body),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: preferenceKeys.display(informationId) });
    },
  });
}

// POST /informations/:id/display/chat
export function useChat(informationId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (body: Parameters<typeof chat>[1]) => chat(informationId, body),
    onSuccess: (data) => {
      if (data.preference_updated) {
        queryClient.invalidateQueries({ queryKey: preferenceKeys.display(informationId) });
      }
    },
  });
}
