package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"knot-api/internal/usecase"

	"uuid"
)

// DisplayController は受信者向けの表示生成・A/B比較・Chatのエンドポイントを提供する。
//
// 受信者(閲覧者)は認証済みのUserとして扱う。recipient_idをリクエストボディで
// クライアントから受け取ることはせず、認証済みJWTのUserIDから解決する。
// 実際に閲覧できるかどうか(access_type、recipients)の判定はusecase層が行う。
type DisplayController struct {
	usecase        *usecase.DisplayUsecase
	accountUsecase *usecase.AccountUsecase
}

// NewDisplayController はDisplayControllerを生成する。
func NewDisplayController(uc *usecase.DisplayUsecase, accountUsecase *usecase.AccountUsecase) *DisplayController {
	return &DisplayController{usecase: uc, accountUsecase: accountUsecase}
}

func parseInformationID(ctx *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid information id"})
		return uuid.Nil(), false
	}
	return id, true
}

// GenerateDisplay はPOST /informations/:id/display を処理する。
// 認証済みの受信者のPreferenceに応じて最適化した表示を返す。
func (c *DisplayController) GenerateDisplay(ctx *gin.Context) {
	informationID, ok := parseInformationID(ctx)
	if !ok {
		return
	}

	_, recipientID, ok := resolveIdentity(ctx, c.accountUsecase)
	if !ok {
		return
	}

	output, err := c.usecase.GenerateDisplay(ctx.Request.Context(), usecase.GenerateDisplayInput{
		InformationID: informationID,
		RecipientID:   recipientID,
	})
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, output)
}

// ListSources はGET /informations/:id/sources を処理する。
// 返答(Response)送信に必要なsource_id・interaction_type・option_idを提示する。
// 返答(SubmitResponse)と同じアクセス制御を適用するため、認証を必須としない
// (PUBLIC+ANONYMOUSなInformationは未ログインでも一覧取得できる)。
func (c *DisplayController) ListSources(ctx *gin.Context) {
	informationID, ok := parseInformationID(ctx)
	if !ok {
		return
	}

	viewerUserID, ok := resolveOptionalIdentity(ctx, c.accountUsecase)
	if !ok {
		return
	}

	views, err := c.usecase.ListSources(ctx.Request.Context(), informationID, viewerUserID)
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, views)
}

type prepareComparisonRequest struct {
	Key    string `json:"key" binding:"required"`
	ValueA string `json:"value_a" binding:"required"`
	ValueB string `json:"value_b" binding:"required"`
}

// PrepareComparison はPOST /informations/:id/display/comparisons を処理する。
// 指定したPreferenceキーの値をA/Bそれぞれに変えた表示を2パターン生成する。
func (c *DisplayController) PrepareComparison(ctx *gin.Context) {
	informationID, ok := parseInformationID(ctx)
	if !ok {
		return
	}

	_, recipientID, ok := resolveIdentity(ctx, c.accountUsecase)
	if !ok {
		return
	}

	var req prepareComparisonRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	output, err := c.usecase.PrepareComparison(ctx.Request.Context(), usecase.PrepareComparisonInput{
		InformationID: informationID,
		RecipientID:   recipientID,
		Comparison: usecase.PreferenceComparison{
			Key:    req.Key,
			ValueA: req.ValueA,
			ValueB: req.ValueB,
		},
	})
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, output)
}

type selectComparisonRequest struct {
	Comparison usecase.PreferenceComparison `json:"comparison" binding:"required"`
	Selected   string                       `json:"selected" binding:"required,oneof=a b"`
}

// SelectComparison はPOST /informations/:id/display/comparisons/selection を処理する。
// 受信者が選んだA/Bの結果をもとに、比較対象だったPreferenceを更新する。
func (c *DisplayController) SelectComparison(ctx *gin.Context) {
	informationID, ok := parseInformationID(ctx)
	if !ok {
		return
	}

	_, recipientID, ok := resolveIdentity(ctx, c.accountUsecase)
	if !ok {
		return
	}

	var req selectComparisonRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.usecase.SelectComparison(ctx.Request.Context(), usecase.SelectComparisonInput{
		InformationID: informationID,
		RecipientID:   recipientID,
		Comparison:    req.Comparison,
		Selected:      req.Selected,
	})
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

type chatRequest struct {
	Messages  []usecase.Message `json:"messages"`
	UserInput string            `json:"user_input" binding:"required"`
}

// Chat はPOST /informations/:id/display/chat を処理する。
// 受信者からの表示調整の要望を1ターン処理する。
func (c *DisplayController) Chat(ctx *gin.Context) {
	informationID, ok := parseInformationID(ctx)
	if !ok {
		return
	}

	_, recipientID, ok := resolveIdentity(ctx, c.accountUsecase)
	if !ok {
		return
	}

	var req chatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	output, err := c.usecase.Chat(ctx.Request.Context(), usecase.ChatInput{
		InformationID: informationID,
		RecipientID:   recipientID,
		Messages:      req.Messages,
		UserInput:     req.UserInput,
	})
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, output)
}
