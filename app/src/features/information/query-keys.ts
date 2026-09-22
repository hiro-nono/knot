export const informationKeys = {
  all: ["informations"] as const,
  mine: () => [...informationKeys.all, "mine"] as const,
  sources: (informationId: string) => [...informationKeys.all, informationId, "sources"] as const,
  recipients: (informationId: string) =>
    [...informationKeys.all, informationId, "recipients"] as const,
};
