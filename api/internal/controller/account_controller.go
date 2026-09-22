package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"knot-api/internal/infrastructure/auth/supabase"
	"knot-api/internal/usecase"

	"uuid"
)

// AccountController はAccount/Userの作成・編集・取得・削除のエンドポイントを提供する。
//
// 呼び出し元(自分/admin)の識別は、認証ミドルウェア(supabase.Middleware)が
// 検証したJWTのUserIDをもとに行う。クライアントが指定したaccount_idを
// そのまま信用することはしない。
type AccountController struct {
	usecase *usecase.AccountUsecase
}

// NewAccountController はAccountControllerを生成する。
func NewAccountController(uc *usecase.AccountUsecase) *AccountController {
	return &AccountController{usecase: uc}
}

// writeError はusecase層から返されたエラーを対応するHTTPステータスへ変換する。
// account/information/display各controllerで共通に使う。
func writeError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrNotFound):
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, usecase.ErrForbidden):
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	default:
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// currentUserID は認証ミドルウェアが設定したClaimsからUserID(ProviderID)を取得する。
// ミドルウェアが適用されていない場合は401を書き込みfalseを返す。
func currentUserID(ctx *gin.Context) (string, bool) {
	claims, ok := supabase.FromContext(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return "", false
	}
	return claims.UserID, true
}

// resolveAccountID は現在の認証ユーザーに対応する内部のAccountIDを解決する。
func (c *AccountController) resolveAccountID(ctx *gin.Context) (uuid.UUID, bool) {
	userID, ok := currentUserID(ctx)
	if !ok {
		return uuid.Nil(), false
	}

	accountID, err := c.usecase.ResolveAccountID(ctx.Request.Context(), userID)
	if err != nil {
		writeError(ctx, err)
		return uuid.Nil(), false
	}

	return accountID, true
}

// resolveIdentity は現在の認証ユーザーに対応する内部のAccountID・UserIDを解決する。
// information/display等、AccountだけでなくUserも必要とするcontrollerが使う。
func resolveIdentity(ctx *gin.Context, accountUsecase *usecase.AccountUsecase) (accountID, userID uuid.UUID, ok bool) {
	providerID, ok := currentUserID(ctx)
	if !ok {
		return uuid.Nil(), uuid.Nil(), false
	}

	accountID, userID, err := accountUsecase.ResolveIdentity(ctx.Request.Context(), providerID)
	if err != nil {
		writeError(ctx, err)
		return uuid.Nil(), uuid.Nil(), false
	}

	return accountID, userID, true
}

// resolveOptionalIdentity は認証情報があれば現在の認証ユーザーに対応するUserIDを
// 解決する。未認証(JWTが無い、またはOptionalMiddlewareで検証をスキップされた)
// 場合はnilを返す(401は書き込まない)。匿名アクセスを許可するエンドポイント
// (response送信、PUBLIC+ANONYMOUSなsource一覧取得)が使う。
// エラー発生時のみokがfalseになり、その場合は呼び出し元でレスポンス書き込み済みなので
// 早期returnすること。
func resolveOptionalIdentity(ctx *gin.Context, accountUsecase *usecase.AccountUsecase) (userID *uuid.UUID, ok bool) {
	claims, authenticated := supabase.FromContext(ctx)
	if !authenticated {
		return nil, true
	}

	_, resolvedUserID, err := accountUsecase.ResolveIdentity(ctx.Request.Context(), claims.UserID)
	if err != nil {
		writeError(ctx, err)
		return nil, false
	}

	return &resolvedUserID, true
}

type registerRequest struct {
	AccountType string  `json:"account_type" binding:"required,oneof=personal organization"`
	Name        *string `json:"name"`
	LastName    string  `json:"last_name" binding:"required"`
	FirstName   string  `json:"first_name" binding:"required"`
	Language    string  `json:"language" binding:"required"`
}

// Register はPOST /accounts を処理する(作成)。
// ProviderIDはリクエストボディではなく、認証済みJWTのUserIDから取得する。
func (c *AccountController) Register(ctx *gin.Context) {
	userID, ok := currentUserID(ctx)
	if !ok {
		return
	}

	var req registerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	view, err := c.usecase.Register(ctx.Request.Context(), usecase.RegisterInput{
		ProviderID:  userID,
		AccountType: req.AccountType,
		Name:        req.Name,
		LastName:    req.LastName,
		FirstName:   req.FirstName,
		Language:    req.Language,
	})
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, view)
}

// GetMe はGET /accounts/me を処理する(自分の取得)。
func (c *AccountController) GetMe(ctx *gin.Context) {
	accountID, ok := c.resolveAccountID(ctx)
	if !ok {
		return
	}

	view, err := c.usecase.GetMe(ctx.Request.Context(), accountID)
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, view)
}

type updateMyProfileRequest struct {
	LastName  string `json:"last_name" binding:"required"`
	FirstName string `json:"first_name" binding:"required"`
	Language  string `json:"language" binding:"required"`
}

// UpdateMyProfile はPATCH /accounts/me を処理する(自分の編集)。
func (c *AccountController) UpdateMyProfile(ctx *gin.Context) {
	accountID, ok := c.resolveAccountID(ctx)
	if !ok {
		return
	}

	var req updateMyProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	view, err := c.usecase.UpdateMyProfile(ctx.Request.Context(), accountID, usecase.UpdateProfileInput{
		LastName:  req.LastName,
		FirstName: req.FirstName,
		Language:  req.Language,
	})
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, view)
}

// ListByStatus はGET /accounts を処理する(ステータスごとの取得、admin専用)。
func (c *AccountController) ListByStatus(ctx *gin.Context) {
	actingAccountID, ok := c.resolveAccountID(ctx)
	if !ok {
		return
	}

	status := ctx.Query("status")
	if status == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "status is required"})
		return
	}

	views, err := c.usecase.ListByStatus(ctx.Request.Context(), actingAccountID, status)
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, views)
}

type updateStatusRequest struct {
	Action string `json:"action" binding:"required"`
}

// UpdateStatus はPATCH /accounts/:id/status を処理する(adminによるアカウント状態の編集)。
func (c *AccountController) UpdateStatus(ctx *gin.Context) {
	targetAccountID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}

	actingAccountID, ok := c.resolveAccountID(ctx)
	if !ok {
		return
	}

	var req updateStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	view, err := c.usecase.UpdateStatus(ctx.Request.Context(), actingAccountID, targetAccountID, usecase.AccountStatusAction(req.Action))
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, view)
}

// Delete はDELETE /accounts/me を処理する(退会、論理削除)。
func (c *AccountController) Delete(ctx *gin.Context) {
	accountID, ok := c.resolveAccountID(ctx)
	if !ok {
		return
	}

	if err := c.usecase.Delete(ctx.Request.Context(), accountID); err != nil {
		writeError(ctx, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// GetUserProfile はGET /users/:id を処理する。
// MembershipView・Recipient一覧などが返すuser_idの表示名(氏名)を解決するために、
// 認証済みユーザーであれば誰の情報でも取得できる(氏名以外は返さない)。
func (c *AccountController) GetUserProfile(ctx *gin.Context) {
	if _, ok := currentUserID(ctx); !ok {
		return
	}

	targetUserID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	view, err := c.usecase.GetUserProfile(ctx.Request.Context(), targetUserID)
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, view)
}
