package controller

import (
	"context"
	"testing"

	"knot-api/internal/usecase"

	"uuid"
)

// newTestAccountUsecase はinformation/display controllerのテストで、
// 認証済みユーザーのAccountID/UserID解決に使うAccountUsecaseを生成する。
func newTestAccountUsecase() *usecase.AccountUsecase {
	return usecase.NewAccountUsecase(
		newInMemoryAccountRepository(),
		newInMemoryUserRepository(),
		newInMemoryMembershipRepository(),
		noopMembershipEventRepository{},
		noopAccountStatusEventRepository{},
		noopTransactionManager{},
	)
}

// registerTestIdentity はaccountUsecase上にAccount/Userを登録し、
// providerIDのJWTから解決されるAccountID/UserIDを返す。
func registerTestIdentity(t *testing.T, accountUsecase *usecase.AccountUsecase, providerID string) (accountID, userID uuid.UUID) {
	t.Helper()

	if _, err := accountUsecase.Register(context.Background(), usecase.RegisterInput{
		ProviderID:  providerID,
		AccountType: "personal",
		LastName:    "山田",
		FirstName:   "太郎",
		Language:    "ja",
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	accountID, userID, err := accountUsecase.ResolveIdentity(context.Background(), providerID)
	if err != nil {
		t.Fatalf("ResolveIdentity() error = %v", err)
	}
	return accountID, userID
}
