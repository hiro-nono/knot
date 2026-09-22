// バックエンド usecase.UserProfileView に対応する。
// GET /users/:id のレスポンス(氏名のみ、メールアドレス等は含まない)。
export interface UserProfileView {
  id: string;
  last_name: string;
  first_name: string;
}
