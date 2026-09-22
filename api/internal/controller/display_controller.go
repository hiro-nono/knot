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

type generateComparisonRequest struct {
	Key string `json:"key" binding:"required"`
}

// GenerateComparison はPOST /informations/:id/display/comparisons を処理する。
// 指定したPreferenceキーについて、AIが決めた対照的な2パターン(A/B)の表示を生成する。
func (c *DisplayController) GenerateComparison(ctx *gin.Context) {
	informationID, ok := parseInformationID(ctx)
	if !ok {
		return
	}

	_, recipientID, ok := resolveIdentity(ctx, c.accountUsecase)
	if !ok {
		return
	}

	var req generateComparisonRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	output, err := c.usecase.GenerateComparison(ctx.Request.Context(), usecase.GenerateComparisonInput{
		InformationID: informationID,
		RecipientID:   recipientID,
		Key:           req.Key,
	})
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, output)
}

type applyPreferenceRequest struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
}

// ApplyPreference はPOST /informations/:id/display/comparisons/apply を処理する。
// GenerateComparisonで確認した候補の値を実際のPreferenceとして永続化する。
func (c *DisplayController) ApplyPreference(ctx *gin.Context) {
	informationID, ok := parseInformationID(ctx)
	if !ok {
		return
	}

	_, recipientID, ok := resolveIdentity(ctx, c.accountUsecase)
	if !ok {
		return
	}

	var req applyPreferenceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.usecase.ApplyPreference(ctx.Request.Context(), usecase.ApplyPreferenceInput{
		InformationID: informationID,
		RecipientID:   recipientID,
		Key:           req.Key,
		Value:         req.Value,
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
// 受信者からの、資料の情報についての質問に1ターン回答する
// (表示の見せ方を調整する機能ではない)。
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
