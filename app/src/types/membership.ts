// バックエンド domain.MembershipRole に対応する。
export type MembershipRole = "owner" | "admin" | "member";

// バックエンド domain.MembershipStatus に対応する。
export type MembershipStatus = "active" | "removed";

// バックエンド domain.MembershipRemovalRequestStatus に対応する。
export type MembershipRemovalRequestStatus = "pending" | "approved" | "rejected";

// バックエンド usecase.MembershipView に対応する。
export interface MembershipView {
  id: string;
  account_id: string;
  user_id: string;
  role: MembershipRole;
  status: MembershipStatus;
  created_at: string;
  updated_at: string;
}

// バックエンド usecase.MembershipRemovalRequestView に対応する。
export interface MembershipRemovalRequestView {
  id: string;
  membership_id: string;
  requested_by_user_id: string;
  status: MembershipRemovalRequestStatus;
  created_at: string;
  updated_at: string;
}

// POST /accounts/:id/members のリクエストボディ。
export interface AddMemberRequest {
  user_id: string;
}

// PATCH /accounts/:id/members/:user_id のリクエストボディ。
// roleは"admin"のみ許可される(owner権限の付与・剥奪はこのAPIでは扱わない)。
export interface GrantAdminRequest {
  role: "admin";
}

// DELETE /accounts/:id/members/:user_id のレスポンス。
// ownerが実行した場合はmembershipが即座に反映され、adminが実行した場合は
// removal_requestとして承認待ちになる。どちらか一方のみ値を持つ。
export type RemoveMemberResponse =
  | { membership: MembershipView; removal_request?: undefined }
  | { membership?: undefined; removal_request: MembershipRemovalRequestView };
