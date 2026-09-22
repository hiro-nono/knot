package usecase

import (
	"context"
	"fmt"

	"knot-api/internal/domain"
	"knot-api/internal/repository"

	"uuid"
)

// ResponseUsecase は受信者がInformationの内容を確認したうえで返答する
// アプリケーションフローを管理する。
//
// 返答できるかどうかは、Recipientそのものではなく、Informationのaccess_type・
// response_policy・recipientsによって判定する(checkResponseEligibility)。
// access_type=publicの場合はresponse_policyに従い、匿名(user_id無し)でも
// 回答できることがある(Google Formのような設計)。access_type=restrictedの
// 場合は常にRecipient+認証済みUserが必須。
type ResponseUsecase struct {
	informationRepo    repository.InformationRepository
	sourceRepo         repository.SourceRepository
	optionRepo         repository.OptionRepository
	recipientRepo      repository.RecipientRepository
	responseRepo       repository.ResponseRepository
	transactionManager repository.TransactionManager
}

// NewResponseUsecase はResponseUsecaseを生成する。
func NewResponseUsecase(
	informationRepo repository.InformationRepository,
	sourceRepo repository.SourceRepository,
	optionRepo repository.OptionRepository,
	recipientRepo repository.RecipientRepository,
	responseRepo repository.ResponseRepository,
	transactionManager repository.TransactionManager,
) *ResponseUsecase {
	return &ResponseUsecase{
		informationRepo:    informationRepo,
		sourceRepo:         sourceRepo,
		optionRepo:         optionRepo,
		recipientRepo:      recipientRepo,
		responseRepo:       responseRepo,
		transactionManager: transactionManager,
	}
}

// ResponseItemInput はSubmitResponseへの入力のうち、1つのSourceに対する
// 回答を表す。OptionIDはcheck・radio用、Valueはtext用で、どちらか一方のみを
// 指定する。
type ResponseItemInput struct {
	SourceID string
	OptionID *string
	Value    *string
}

// SubmitResponseInput はSubmitResponseへの入力。
// UserIDはnil(未認証・匿名)を許容する。InformationのaccessType・
// responsePolicyがそれを許可している場合のみ、nilのままの送信が受理される。
type SubmitResponseInput struct {
	InformationID uuid.UUID
	UserID        *uuid.UUID
	Items         []ResponseItemInput
}

// SubmitResponse は受信者がInformationに対する返答を新規登録する。
// 返答できるかどうかは、InformationのRecipientそのものではなくaccess_type・
// response_policy・recipientsによって判定する(checkResponseEligibility)。
// 各回答は、対象のSourceのinteraction_typeに応じてoption_id(radio・check)
// またはvalue(text)のいずれかである必要があり、radio・checkの場合はそのSourceに
// 属するOptionである必要がある。
func (u *ResponseUsecase) SubmitResponse(ctx context.Context, input SubmitResponseInput) (*ResponseView, error) {
	information, err := u.informationRepo.Get(ctx, input.InformationID)
	if err != nil {
		return nil, fmt.Errorf("get information: %w", err)
	}

	if err := u.checkResponseEligibility(ctx, information, input.UserID); err != nil {
		return nil, err
	}

	sources, err := u.loadSources(ctx, input.InformationID)
	if err != nil {
		return nil, err
	}

	response, err := domain.NewResponse(input.InformationID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("build response: %w", err)
	}

	for _, itemInput := range input.Items {
		item, err := buildResponseItem(response.ID(), sources, itemInput)
		if err != nil {
			return nil, err
		}
		if err := response.AddItem(*item); err != nil {
			return nil, fmt.Errorf("add response item: %w", err)
		}
	}

	if err := u.transactionManager.WithinTransaction(ctx, func(ctx context.Context) error {
		return u.responseRepo.Create(ctx, response)
	}); err != nil {
		return nil, err
	}

	return newResponseView(response), nil
}

// checkResponseEligibility は、userID(未認証の場合はnil)がinformationに対して
// 返答できるかどうかを判定する。組み合わせごとの判定ルールは次のとおり:
//   - restricted:              常に認証済みUser(userID != nil)かつRecipientである必要がある
//   - public + authenticated:  認証済みUserである必要がある(Recipientである必要はない)
//   - public + anonymous:      userIDの有無を問わず誰でも回答できる
func (u *ResponseUsecase) checkResponseEligibility(ctx context.Context, information *domain.Information, userID *uuid.UUID) error {
	allowed, err := canRespondToInformation(ctx, u.recipientRepo, information, userID)
	if err != nil {
		return err
	}
	if !allowed {
		return domain.ErrForbidden
	}
	return nil
}

// ListResponses はInformationに寄せられたResponseを一覧取得する
// (そのInformationを所有するAccountのみが実行できる)。
func (u *ResponseUsecase) ListResponses(ctx context.Context, actingAccountID, informationID uuid.UUID) ([]*ResponseView, error) {
	information, err := u.informationRepo.Get(ctx, informationID)
	if err != nil {
		return nil, fmt.Errorf("get information: %w", err)
	}
	if information.AccountID() != actingAccountID {
		return nil, domain.ErrForbidden
	}

	responses, err := u.responseRepo.ListByInformationID(ctx, informationID)
	if err != nil {
		return nil, fmt.Errorf("list responses: %w", err)
	}

	views := make([]*ResponseView, 0, len(responses))
	for _, response := range responses {
		views = append(views, newResponseView(response))
	}

	return views, nil
}

// buildResponseItem はitemInputを検証し、domain.ResponseItemを構築する。
// sourceIDが対象のInformationに属していること、interaction_typeに応じて
// option_id/valueが正しく指定されていること(radio・checkはoption_idのみ、
// textはvalueのみ)、option_idを指定する場合はそのSourceに属するOptionで
// あることを検証する。
func buildResponseItem(responseID uuid.UUID, sources map[uuid.UUID]*domain.Source, itemInput ResponseItemInput) (*domain.ResponseItem, error) {
	sourceID, err := uuid.Parse(itemInput.SourceID)
	if err != nil {
		return nil, fmt.Errorf("parse source id: %w", err)
	}

	source, ok := sources[sourceID]
	if !ok {
		return nil, fmt.Errorf("source %s は指定されたinformationに属していません", sourceID)
	}

	var optionID *uuid.UUID
	if itemInput.OptionID != nil {
		parsed, err := uuid.Parse(*itemInput.OptionID)
		if err != nil {
			return nil, fmt.Errorf("parse option id: %w", err)
		}
		optionID = &parsed
	}

	if err := validateResponseItemAgainstSource(source, optionID, itemInput.Value); err != nil {
		return nil, err
	}

	return domain.NewResponseItem(responseID, sourceID, optionID, itemInput.Value)
}

// validateResponseItemAgainstSource はsourceのinteraction_typeと
// optionID・valueの組み合わせが整合しているかを検証する。
func validateResponseItemAgainstSource(source *domain.Source, optionID *uuid.UUID, value *string) error {
	interactionType := source.InteractionType()
	if interactionType == nil {
		return fmt.Errorf("source %s は返答を必要としません", source.ID())
	}

	switch *interactionType {
	case domain.SourceInteractionTypeRadio, domain.SourceInteractionTypeCheck:
		if optionID == nil {
			return fmt.Errorf("source %s はoption_idの指定が必須です", source.ID())
		}
		if value != nil {
			return fmt.Errorf("source %s にvalueを指定することはできません", source.ID())
		}
		for _, option := range source.Options() {
			if option.ID() == *optionID {
				return nil
			}
		}
		return fmt.Errorf("option_id %s はsource %s に属していません", *optionID, source.ID())
	case domain.SourceInteractionTypeText:
		if value == nil {
			return fmt.Errorf("source %s はvalueの指定が必須です", source.ID())
		}
		if optionID != nil {
			return fmt.Errorf("source %s にoption_idを指定することはできません", source.ID())
		}
		return nil
	default:
		return fmt.Errorf("source %s のinteraction_typeが不明です: %q", source.ID(), *interactionType)
	}
}

// loadSources はinformationIDに属するSourceを、配下のOptionまで含めて取得する。
func (u *ResponseUsecase) loadSources(ctx context.Context, informationID uuid.UUID) (map[uuid.UUID]*domain.Source, error) {
	sources, err := u.sourceRepo.List(ctx, informationID)
	if err != nil {
		return nil, fmt.Errorf("list sources: %w", err)
	}

	result := make(map[uuid.UUID]*domain.Source, len(sources))
	for _, source := range sources {
		options, err := u.optionRepo.List(ctx, source.ID())
		if err != nil {
			return nil, fmt.Errorf("list options: %w", err)
		}
		for _, option := range options {
			if err := source.AddOption(*option); err != nil {
				return nil, fmt.Errorf("add option: %w", err)
			}
		}
		result[source.ID()] = source
	}

	return result, nil
}
