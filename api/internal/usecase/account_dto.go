package usecase

import (
	"time"

	"knot-api/internal/domain"
)

// AccountView はAccountとUserを組み合わせた、外部に返却するための表現。
type AccountView struct {
	ID          string    `json:"id"`
	ProviderID  string    `json:"provider_id"`
	AccountType string    `json:"account_type"`
	Name        *string   `json:"name,omitempty"`
	Role        string    `json:"role"`
	Status      string    `json:"status"`
	UserID      string    `json:"user_id"`
	LastName    string    `json:"last_name"`
	FirstName   string    `json:"first_name"`
	Language    string    `json:"language"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserProfileView はUserの最小限のプロフィール(氏名)を外部に返却するための表現。
// メンバー一覧・recipient一覧などでuser_idから表示名を解決するために使う
// (メールアドレス等の機微な情報は含めない)。
type UserProfileView struct {
	ID        string `json:"id"`
	LastName  string `json:"last_name"`
	FirstName string `json:"first_name"`
}

func newUserProfileView(user *domain.User) *UserProfileView {
	return &UserProfileView{
		ID:        user.ID().String(),
		LastName:  user.LastName(),
		FirstName: user.FirstName(),
	}
}

func newAccountView(account *domain.Account, user *domain.User) *AccountView {
	return &AccountView{
		ID:          account.ID().String(),
		ProviderID:  account.ProviderID(),
		AccountType: string(account.AccountType()),
		Name:        account.Name(),
		Role:        string(account.Role()),
		Status:      string(account.Status()),
		UserID:      user.ID().String(),
		LastName:    user.LastName(),
		FirstName:   user.FirstName(),
		Language:    user.Language(),
		CreatedAt:   account.CreatedAt(),
		UpdatedAt:   account.UpdatedAt(),
	}
}
