import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { listResponses, submitResponse } from "@/features/response/api";
import { responseKeys } from "@/features/response/query-keys";

// GET /informations/:id/responses
export function useResponses(informationId: string, enabled = true) {
  return useQuery({
    queryKey: responseKeys.byInformation(informationId),
    queryFn: () => listResponses(informationId),
    enabled,
  });
}

// POST /informations/:id/responses
export function useSubmitResponse(informationId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (body: Parameters<typeof submitResponse>[1]) => submitResponse(informationId, body),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: responseKeys.byInformation(informationId) });
    },
  });
}
