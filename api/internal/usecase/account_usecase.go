package usecase

import (
	"context"
	"fmt"

	"knot-api/internal/domain"
	"knot-api/internal/repository"

	"uuid"
)

// AccountUsecase はAccount/Userの作成・編集・取得・削除のアプリケーションフローを管理する。
//
// 認証(Supabaseが発行したトークンの検証)自体はこのUseCaseの責務ではなく、
// 呼び出し元(controller)がすでに検証済みのaccountIDを渡してくることを前提とする。
// admin専用の操作は、actingAccountIDのRoleを見て許可判定する。
type AccountUsecase struct {
	accountRepo         repository.AccountRepository
	userRepo            repository.UserRepository
	membershipRepo      repository.MembershipRepository
	membershipEventRepo repository.MembershipEventRepository
	statusEventRepo     repository.AccountStatusEventRepository
	transactionManager  repository.TransactionManager
}

// NewAccountUsecase はAccountUsecaseを生成する。
func NewAccountUsecase(
	accountRepo repository.AccountRepository,
	userRepo repository.UserRepository,
	membershipRepo repository.MembershipRepository,
	membershipEventRepo repository.MembershipEventRepository,
	statusEventRepo repository.AccountStatusEventRepository,
	transactionManager repository.TransactionManager,
) *AccountUsecase {
	return &AccountUsecase{
		accountRepo:         accountRepo,
		userRepo:            userRepo,
		membershipRepo:      membershipRepo,
		membershipEventRepo: membershipEventRepo,
		statusEventRepo:     statusEventRepo,
		transactionManager:  transactionManager,
	}
}

// RegisterInput はRegisterへの入力。
// Nameは、AccountTypeがorganizationの場合のみ指定する(personalの場合はnilである必要がある)。
type RegisterInput struct {
	ProviderID  string
	AccountType string
	Name        *string
	LastName    string
	FirstName   string
	Language    string
}

// Register はOrganization(Account、role=user、status=active)を新規作成し、
// 最初に登録されたUserを自動的にownerとしてMembershipへ登録する。
func (u *AccountUsecase) Register(ctx context.Context, input RegisterInput) (*AccountView, error) {
	account, err := domain.NewAccount(input.ProviderID, domain.AccountType(input.AccountType), input.Name)
	if err != nil {
		return nil, fmt.Errorf("build account: %w", err)
	}

	user, err := domain.NewUser(input.LastName, input.FirstName, input.Language)
	if err != nil {
		return nil, fmt.Errorf("build user: %w", err)
	}

	membership, err := domain.NewMembership(account.ID(), user.ID(), domain.MembershipRoleOwner)
	if err != nil {
		return nil, fmt.Errorf("build membership: %w", err)
	}

	membershipEvent, err := domain.NewMembershipEvent(membership.ID(), domain.MembershipEventCreated)
	if err != nil {
		return nil, fmt.Errorf("build membership event: %w", err)
	}

	event, err := domain.NewAccountStatusEvent(account.ID(), domain.AccountStatusEventCreated)
	if err != nil {
		return nil, fmt.Errorf("build status event: %w", err)
	}

	err = u.transactionManager.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := u.accountRepo.Create(ctx, account); err != nil {
			return fmt.Errorf("create account: %w", err)
		}
		if err := u.userRepo.Create(ctx, user); err != nil {
			return fmt.Errorf("create user: %w", err)
		}
		if err := u.membershipRepo.Create(ctx, membership); err != nil {
			return fmt.Errorf("create membership: %w", err)
		}
		if err := u.membershipEventRepo.Create(ctx, membershipEvent); err != nil {
			return fmt.Errorf("create membership event: %w", err)
		}
		if err := u.statusEventRepo.Create(ctx, event); err != nil {
			return fmt.Errorf("create status event: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return newAccountView(account, user), nil
}

// GetMe はaccountIDに対応するAccount/Userを取得する(自分の取得)。
func (u *AccountUsecase) GetMe(ctx context.Context, accountID uuid.UUID) (*AccountView, error) {
	return u.getView(ctx, accountID)
}

// UpdateProfileInput はUpdateMyProfileへの入力。
type UpdateProfileInput struct {
	LastName  string
	FirstName string
	Language  string
}

// UpdateMyProfile は本人がUserのプロフィール(姓・名・言語)を編集する(自分の編集)。
func (u *AccountUsecase) UpdateMyProfile(ctx context.Context, accountID uuid.UUID, input UpdateProfileInput) (*AccountView, error) {
	account, err := u.accountRepo.Get(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("get account: %w", err)
	}

	user, err := u.getOwnerUser(ctx, accountID)
	if err != nil {
		return nil, err
	}

	if err := user.UpdateProfile(input.LastName, input.FirstName, input.Language); err != nil {
		return nil, fmt.Errorf("update profile: %w", err)
	}

	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("save user: %w", err)
	}

	return newAccountView(account, user), nil
}

// ListByStatus はstatusを指定してAccount一覧を取得する(admin専用)。
// withdrawn(退会)は対象外とする。
func (u *AccountUsecase) ListByStatus(ctx context.Context, actingAccountID uuid.UUID, status string) ([]*AccountView, error) {
	accountStatus := domain.AccountStatus(status)
	if accountStatus == domain.AccountStatusWithdrawn {
		return nil, fmt.Errorf("listing withdrawn accounts is not allowed")
	}

	if _, err := u.requireAdmin(ctx, actingAccountID); err != nil {
		return nil, err
	}

	accounts, err := u.accountRepo.ListByStatus(ctx, accountStatus)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}

	views := make([]*AccountView, 0, len(accounts))
	for _, account := range accounts {
		user, err := u.getOwnerUser(ctx, account.ID())
		if err != nil {
			return nil, err
		}
		views = append(views, newAccountView(account, user))
	}

	return views, nil
}

// AccountStatusAction はUpdateStatusで指定する操作の種別を表す。
type AccountStatusAction string

const (
	AccountStatusActionFreeze     AccountStatusAction = "freeze"
	AccountStatusActionSuspend    AccountStatusAction = "suspend"
	AccountStatusActionBan        AccountStatusAction = "ban"
	AccountStatusActionUnfreeze   AccountStatusAction = "unfreeze"
	AccountStatusActionReactivate AccountStatusAction = "reactivate"
)

// UpdateStatus はadminが対象Accountのステータスを更新する(adminによるアカウント状態の編集)。
// 更新は履歴(AccountStatusEvent)への記録とあわせて1つのトランザクションで行う。
func (u *AccountUsecase) UpdateStatus(ctx context.Context, actingAccountID uuid.UUID, targetAccountID uuid.UUID, action AccountStatusAction) (*AccountView, error) {
	if _, err := u.requireAdmin(ctx, actingAccountID); err != nil {
		return nil, err
	}

	target, err := u.accountRepo.Get(ctx, targetAccountID)
	if err != nil {
		return nil, fmt.Errorf("get target account: %w", err)
	}

	eventType, err := applyAccountStatusAction(target, action)
	if err != nil {
		return nil, fmt.Errorf("apply status action: %w", err)
	}

	event, err := domain.NewAccountStatusEvent(target.ID(), eventType)
	if err != nil {
		return nil, fmt.Errorf("build status event: %w", err)
	}

	err = u.transactionManager.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := u.accountRepo.Update(ctx, target); err != nil {
			return fmt.Errorf("update account: %w", err)
		}
		if err := u.statusEventRepo.Create(ctx, event); err != nil {
			return fmt.Errorf("create status event: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	user, err := u.getOwnerUser(ctx, target.ID())
	if err != nil {
		return nil, err
	}

	return newAccountView(target, user), nil
}

func applyAccountStatusAction(account *domain.Account, action AccountStatusAction) (domain.AccountStatusEventType, error) {
	switch action {
	case AccountStatusActionFreeze:
		return account.Freeze()
	case AccountStatusActionSuspend:
		return account.Suspend()
	case AccountStatusActionBan:
		return account.Ban()
	case AccountStatusActionUnfreeze:
		return account.Unfreeze()
	case AccountStatusActionReactivate:
		return account.Reactivate()
	default:
		return "", fmt.Errorf("unknown action %q", action)
	}
}

// Delete は本人のAccountを退会(論理削除)する。
// statusをwithdrawnに更新し、履歴(AccountStatusEvent)への記録とあわせて
// 1つのトランザクションで行う。
func (u *AccountUsecase) Delete(ctx context.Context, accountID uuid.UUID) error {
	account, err := u.accountRepo.Get(ctx, accountID)
	if err != nil {
		return fmt.Errorf("get account: %w", err)
	}

	eventType, err := account.Withdraw()
	if err != nil {
		return fmt.Errorf("withdraw account: %w", err)
	}

	event, err := domain.NewAccountStatusEvent(account.ID(), eventType)
	if err != nil {
		return fmt.Errorf("build status event: %w", err)
	}

	return u.transactionManager.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := u.accountRepo.Update(ctx, account); err != nil {
			return fmt.Errorf("update account: %w", err)
		}
		if err := u.statusEventRepo.Create(ctx, event); err != nil {
			return fmt.Errorf("create status event: %w", err)
		}
		return nil
	})
}

// ResolveAccountID は外部IDプロバイダ(Supabase)のユーザーIDから、
// 対応するAccountのIDを取得する。認証ミドルウェアが検証したUserIDを
// controllerが内部のaccountIDに変換するために使う。
func (u *AccountUsecase) ResolveAccountID(ctx context.Context, providerID string) (uuid.UUID, error) {
	account, err := u.accountRepo.GetByProviderID(ctx, providerID)
	if err != nil {
		return uuid.Nil(), fmt.Errorf("get account by provider id: %w", err)
	}
	return account.ID(), nil
}

// ResolveIdentity は外部IDプロバイダ(Supabase)のユーザーIDから、
// 対応する内部のAccountID・UserIDを取得する。
// controllerが認証済みユーザーとして操作を行う際に、Account/Userの両方を
// 必要とするusecase(Information/Display等)への橋渡しに使う。
func (u *AccountUsecase) ResolveIdentity(ctx context.Context, providerID string) (accountID uuid.UUID, userID uuid.UUID, err error) {
	account, err := u.accountRepo.GetByProviderID(ctx, providerID)
	if err != nil {
		return uuid.Nil(), uuid.Nil(), fmt.Errorf("get account by provider id: %w", err)
	}

	user, err := u.getOwnerUser(ctx, account.ID())
	if err != nil {
		return uuid.Nil(), uuid.Nil(), err
	}

	return account.ID(), user.ID(), nil
}

// GetUserProfile はuserIDに対応するUserの最小限のプロフィール(氏名)を取得する。
// MembershipView・Recipient一覧などが返すuser_idから表示名を解決するために、
// 認証済みユーザーであれば誰でも呼び出せる(氏名以外の機微な情報は返さない)。
func (u *AccountUsecase) GetUserProfile(ctx context.Context, userID uuid.UUID) (*UserProfileView, error) {
	user, err := u.userRepo.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return newUserProfileView(user), nil
}

func (u *AccountUsecase) getView(ctx context.Context, accountID uuid.UUID) (*AccountView, error) {
	account, err := u.accountRepo.Get(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("get account: %w", err)
	}

	user, err := u.getOwnerUser(ctx, accountID)
	if err != nil {
		return nil, err
	}

	return newAccountView(account, user), nil
}

// getOwnerUser はaccountIDに対応するowner roleのMembershipからUserを解決する。
// Accountはowner Membershipを持つUserによって作成されるため、
// 「そのAccountの持ち主」としてのUserプロフィールはこの経路で取得する。
func (u *AccountUsecase) getOwnerUser(ctx context.Context, accountID uuid.UUID) (*domain.User, error) {
	membership, err := u.membershipRepo.GetOwnerByAccountID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("get owner membership: %w", err)
	}

	user, err := u.userRepo.Get(ctx, membership.UserID())
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	return user, nil
}

func (u *AccountUsecase) requireAdmin(ctx context.Context, accountID uuid.UUID) (*domain.Account, error) {
	account, err := u.accountRepo.Get(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("get acting account: %w", err)
	}
	if !account.IsAdmin() {
		return nil, domain.ErrForbidden
	}
	return account, nil
}
