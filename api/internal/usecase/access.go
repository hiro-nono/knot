package usecase

import (
	"context"
	"fmt"

	"knot-api/internal/domain"
	"knot-api/internal/repository"

	"uuid"
)

// canAccessInformation はviewerUserIDがinformationを閲覧・操作できるかどうかを
// 判定する。access_typeがpublicであればリンクを知っている全員が対象になり、
// restrictedであればrecipientsに登録されたUserのみが対象になる。
//
// 閲覧(Display)・返答(Response)など、Informationへのアクセス可否判定を
// 必要とする複数のUseCaseから共通して使う。
func canAccessInformation(ctx context.Context, recipientRepo repository.RecipientRepository, information *domain.Information, viewerUserID uuid.UUID) (bool, error) {
	if information.IsPublic() {
		return true, nil
	}

	exists, err := recipientRepo.Exists(ctx, information.ID(), viewerUserID)
	if err != nil {
		return false, fmt.Errorf("check recipient: %w", err)
	}

	return exists, nil
}

// canRespondToInformation はuserID(未認証の場合はnil)がinformationに対して
// 返答、またはその前提となるSource一覧の閲覧を行えるかどうかを判定する。
// ResponseUsecase.SubmitResponseとDisplayUsecase.ListSourcesの両方が使う。
//
// 常に認証済みUserを要求するcanAccessInformation(表示生成・Chat等の閲覧用)とは
// 別の判定基準であり、access_type・response_policyの組み合わせに応じて
// 未認証(userID == nil)でも許可することがある。
//   - restricted:              常に認証済みUserかつRecipientである必要がある
//   - public + authenticated:  認証済みUserである必要がある(Recipientである必要はない)
//   - public + anonymous:      userIDの有無を問わず誰でも許可される
func canRespondToInformation(ctx context.Context, recipientRepo repository.RecipientRepository, information *domain.Information, userID *uuid.UUID) (bool, error) {
	if information.RequiresAuthenticatedResponder() && userID == nil {
		return false, nil
	}

	if information.RequiresRecipientForResponse() {
		exists, err := recipientRepo.Exists(ctx, information.ID(), *userID)
		if err != nil {
			return false, fmt.Errorf("check recipient: %w", err)
		}
		return exists, nil
	}

	return true, nil
}
