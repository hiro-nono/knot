export const userKeys = {
  all: ["users"] as const,
  profile: (userId: string) => [...userKeys.all, userId] as const,
};
