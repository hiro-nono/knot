package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"knot-api/internal/usecase"
)

// ResponseController は受信者がInformationに対して返答するエンドポイント、
// および発信者がその返答を確認するエンドポイントを提供する。
type ResponseController struct {
	usecase        *usecase.ResponseUsecase
	accountUsecase *usecase.AccountUsecase
}

// NewResponseController はResponseControllerを生成する。
func NewResponseController(uc *usecase.ResponseUsecase, accountUsecase *usecase.AccountUsecase) *ResponseController {
	return &ResponseController{usecase: uc, accountUsecase: accountUsecase}
}

type submitResponseItemRequest struct {
	SourceID string  `json:"source_id" binding:"required"`
	OptionID *string `json:"option_id"`
	Value    *string `json:"value"`
}

type submitResponseRequest struct {
	Items []submitResponseItemRequest `json:"items" binding:"required,min=1,dive"`
}

// SubmitResponse はPOST /informations/:id/responses を処理する。
// 受信者がInformationに対する返答を登録する。このエンドポイントは認証を必須と
// しない(supabase.OptionalMiddlewareを使う)。PUBLIC+ANONYMOUSなInformationは
// 未ログインでも回答できるため。JWTが送られていれば、解決したUserIDを使う
// (匿名回答を許可するInformationであっても、ログイン済みならその情報を記録する)。
// 実際に回答できるかどうか(認証・Recipientの要否)の判定はusecase層が行う。
func (c *ResponseController) SubmitResponse(ctx *gin.Context) {
	informationID, ok := parseInformationID(ctx)
	if !ok {
		return
	}

	userID, ok := resolveOptionalIdentity(ctx, c.accountUsecase)
	if !ok {
		return
	}

	var req submitResponseRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	items := make([]usecase.ResponseItemInput, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, usecase.ResponseItemInput{
			SourceID: item.SourceID,
			OptionID: item.OptionID,
			Value:    item.Value,
		})
	}

	view, err := c.usecase.SubmitResponse(ctx.Request.Context(), usecase.SubmitResponseInput{
		InformationID: informationID,
		UserID:        userID,
		Items:         items,
	})
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, view)
}

// ListResponses はGET /informations/:id/responses を処理する。
// Informationに寄せられたResponseを一覧取得する(そのInformationを所有する
// Accountのみが実行できる)。
func (c *ResponseController) ListResponses(ctx *gin.Context) {
	informationID, ok := parseInformationID(ctx)
	if !ok {
		return
	}

	actingAccountID, _, ok := resolveIdentity(ctx, c.accountUsecase)
	if !ok {
		return
	}

	views, err := c.usecase.ListResponses(ctx.Request.Context(), actingAccountID, informationID)
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, views)
}
