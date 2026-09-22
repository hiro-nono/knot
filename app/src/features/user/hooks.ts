import { useQuery } from "@tanstack/react-query";

import { getUserProfile } from "@/features/user/api";
import { userKeys } from "@/features/user/query-keys";

// GET /users/:id
export function useUserProfile(userId: string, enabled = true) {
  return useQuery({
    queryKey: userKeys.profile(userId),
    queryFn: () => getUserProfile(userId),
    enabled: enabled && Boolean(userId),
  });
}
