export const responseKeys = {
  all: ["responses"] as const,
  byInformation: (informationId: string) => [...responseKeys.all, informationId] as const,
};
