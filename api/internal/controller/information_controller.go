package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"knot-api/internal/usecase"

	"uuid"
)

// InformationController は発信者とAIの対話エンドポイント、および
// Informationの共有設定(recipients)を管理するエンドポイントを提供する。
type InformationController struct {
	usecase        *usecase.InformationUsecase
	accountUsecase *usecase.AccountUsecase
}

// NewInformationController はInformationControllerを生成する。
func NewInformationController(uc *usecase.InformationUsecase, accountUsecase *usecase.AccountUsecase) *InformationController {
	return &InformationController{usecase: uc, accountUsecase: accountUsecase}
}

// ListMine はGET /informations を処理する。
// 認証済みの発信者が過去に作成したInformationを一覧取得する。
func (c *InformationController) ListMine(ctx *gin.Context) {
	accountID, _, ok := resolveIdentity(ctx, c.accountUsecase)
	if !ok {
		return
	}

	views, err := c.usecase.ListMine(ctx.Request.Context(), accountID)
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, views)
}

type processInformationRequest struct {
	AccessType     string            `json:"access_type" binding:"required,oneof=public restricted"`
	ResponsePolicy string            `json:"response_policy" binding:"required,oneof=anonymous authenticated"`
	Messages       []usecase.Message `json:"messages"`
	UserInput      string            `json:"user_input" binding:"required"`
}

// ProcessInformation は発信者との対話を1ターン進める。
// リクエストで受け取った会話履歴と発言をUseCaseへ渡し、
// AIの応答(質問または確定結果)をそのまま返却する。
// 発信者は認証済みである必要があり、確定時のInformationはそのAccount/Userに
// 紐づけて永続化される。
func (c *InformationController) ProcessInformation(ctx *gin.Context) {
	accountID, userID, ok := resolveIdentity(ctx, c.accountUsecase)
	if !ok {
		return
	}

	var req processInformationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	output, err := c.usecase.ProcessInformation(ctx.Request.Context(), usecase.ProcessInformationInput{
		AccountID:       accountID,
		CreatedByUserID: userID,
		AccessType:      req.AccessType,
		ResponsePolicy:  req.ResponsePolicy,
		Messages:        req.Messages,
		UserInput:       req.UserInput,
	})
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, output)
}

type addRecipientRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

// AddRecipient はPOST /informations/:id/recipients を処理する。
// restrictedなInformationを閲覧できるUserを1件追加する(そのInformationを
// 所有するAccountのみが実行できる)。
func (c *InformationController) AddRecipient(ctx *gin.Context) {
	informationID, ok := parseInformationID(ctx)
	if !ok {
		return
	}

	actingAccountID, _, ok := resolveIdentity(ctx, c.accountUsecase)
	if !ok {
		return
	}

	var req addRecipientRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	recipientUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	err = c.usecase.AddRecipient(ctx.Request.Context(), usecase.RecipientInput{
		ActingAccountID: actingAccountID,
		InformationID:   informationID,
		UserID:          recipientUserID,
	})
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// RemoveRecipient はDELETE /informations/:id/recipients/:user_id を処理する。
// Informationの閲覧許可を取り消す(そのInformationを所有するAccountのみが実行できる)。
func (c *InformationController) RemoveRecipient(ctx *gin.Context) {
	informationID, ok := parseInformationID(ctx)
	if !ok {
		return
	}

	recipientUserID, err := uuid.Parse(ctx.Param("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	actingAccountID, _, ok := resolveIdentity(ctx, c.accountUsecase)
	if !ok {
		return
	}

	err = c.usecase.RemoveRecipient(ctx.Request.Context(), usecase.RecipientInput{
		ActingAccountID: actingAccountID,
		InformationID:   informationID,
		UserID:          recipientUserID,
	})
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// ListRecipients はGET /informations/:id/recipients を処理する。
// Informationに登録されているRecipient(閲覧可能なUserのID)を一覧取得する
// (そのInformationを所有するAccountのみが実行できる)。
func (c *InformationController) ListRecipients(ctx *gin.Context) {
	informationID, ok := parseInformationID(ctx)
	if !ok {
		return
	}

	actingAccountID, _, ok := resolveIdentity(ctx, c.accountUsecase)
	if !ok {
		return
	}

	userIDs, err := c.usecase.ListRecipients(ctx.Request.Context(), actingAccountID, informationID)
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"user_ids": userIDs})
}
