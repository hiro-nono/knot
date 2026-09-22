export const preferenceKeys = {
  all: ["preference"] as const,
  display: (informationId: string) => [...preferenceKeys.all, informationId, "display"] as const,
};
