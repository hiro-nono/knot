export const membershipKeys = {
  all: ["memberships"] as const,
  members: (accountId: string) => [...membershipKeys.all, accountId, "members"] as const,
  removalRequests: (accountId: string) =>
    [...membershipKeys.all, accountId, "removal-requests"] as const,
};
