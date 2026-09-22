import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import {
  addRecipient,
  listMyInformations,
  listRecipients,
  listSources,
  processInformation,
  removeRecipient,
} from "@/features/information/api";
import { informationKeys } from "@/features/information/query-keys";

// GET /informations
export function useMyInformations(enabled = true) {
  return useQuery({
    queryKey: informationKeys.mine(),
    queryFn: listMyInformations,
    enabled,
  });
}

// POST /informations/messages
// 対話1ターンごとに呼び出す(確定するまで呼び出し元がmessagesを積み上げて渡す)。
export function useProcessInformation() {
  return useMutation({
    mutationFn: processInformation,
  });
}

// GET /informations/:id/sources
export function useSources(informationId: string, enabled = true) {
  return useQuery({
    queryKey: informationKeys.sources(informationId),
    queryFn: () => listSources(informationId),
    enabled,
  });
}

// GET /informations/:id/recipients
export function useRecipients(informationId: string, enabled = true) {
  return useQuery({
    queryKey: informationKeys.recipients(informationId),
    queryFn: () => listRecipients(informationId),
    enabled,
  });
}

// POST /informations/:id/recipients
export function useAddRecipient(informationId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (userId: string) => addRecipient(informationId, userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: informationKeys.recipients(informationId) });
    },
  });
}

// DELETE /informations/:id/recipients/:user_id
export function useRemoveRecipient(informationId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (userId: string) => removeRecipient(informationId, userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: informationKeys.recipients(informationId) });
    },
  });
}
