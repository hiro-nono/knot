// バックエンド domain.AccountType に対応する。
export type AccountType = "personal" | "organization";

// バックエンド domain.AccountRole に対応する。
export type AccountRole = "admin" | "user";

// バックエンド domain.AccountStatus に対応する。
// withdrawnはListByStatusの対象外(バックエンドが拒否する)。
export type AccountStatus = "active" | "frozen" | "suspended" | "withdrawn" | "banned";

// バックエンド usecase.AccountStatusAction に対応する。
export type AccountStatusAction =
  | "freeze"
  | "suspend"
  | "ban"
  | "unfreeze"
  | "reactivate";

// バックエンド usecase.AccountView に対応する。
export interface AccountView {
  id: string;
  provider_id: string;
  account_type: AccountType;
  name?: string;
  role: AccountRole;
  status: AccountStatus;
  user_id: string;
  last_name: string;
  first_name: string;
  language: string;
  created_at: string;
  updated_at: string;
}

// POST /accounts のリクエストボディ。
// nameはaccount_typeがorganizationの場合のみ指定する。
export interface RegisterAccountRequest {
  account_type: AccountType;
  name?: string | null;
  last_name: string;
  first_name: string;
  language: string;
}

// PATCH /accounts/me のリクエストボディ。
export interface UpdateMyProfileRequest {
  last_name: string;
  first_name: string;
  language: string;
}

// PATCH /accounts/:id/status のリクエストボディ。
export interface UpdateAccountStatusRequest {
  action: AccountStatusAction;
}
