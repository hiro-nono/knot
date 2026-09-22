import { apiClient } from "@/lib/api/client";
import type { UserProfileView } from "@/types/user";

// GET /users/:id
// メンバー一覧・recipient一覧などが返すuser_idから表示名(氏名)を解決するために使う。
export function getUserProfile(userId: string) {
  return apiClient.get<UserProfileView>(`/users/${userId}`);
}
