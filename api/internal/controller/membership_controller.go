package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"knot-api/internal/usecase"

	"uuid"
)

// MembershipController はOrganization(Account)内でのUserの所属・権限管理
// (追加・ADMIN権限付与・除外・除外申請の承認/却下)のエンドポイントを提供する。
//
// 権限判定はすべてusecase層(MembershipUsecase)で行う。このcontrollerは
// クライアントが送信したaccount_id・roleを信頼せず、JWTから解決した
// actingUserIDのみをusecaseへ渡す。
type MembershipController struct {
	usecase        *usecase.MembershipUsecase
	accountUsecase *usecase.AccountUsecase
}

// NewMembershipController はMembershipControllerを生成する。
func NewMembershipController(uc *usecase.MembershipUsecase, accountUsecase *usecase.AccountUsecase) *MembershipController {
	return &MembershipController{usecase: uc, accountUsecase: accountUsecase}
}

func parseTargetAccountID(ctx *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return uuid.Nil(), false
	}
	return id, true
}

type addMemberRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

// AddMember はPOST /accounts/:id/members を処理する。
// Organizationへ新しいUserを追加する(owner専用、初期roleはmember)。
func (c *MembershipController) AddMember(ctx *gin.Context) {
	targetAccountID, ok := parseTargetAccountID(ctx)
	if !ok {
		return
	}

	_, actingUserID, ok := resolveIdentity(ctx, c.accountUsecase)
	if !ok {
		return
	}

	var req addMemberRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	targetUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	view, err := c.usecase.AddUser(ctx.Request.Context(), usecase.AddUserInput{
		ActingUserID:    actingUserID,
		TargetAccountID: targetAccountID,
		TargetUserID:    targetUserID,
	})
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, view)
}

type grantAdminRequest struct {
	Role string `json:"role" binding:"required,oneof=admin"`
}

// GrantAdmin はPATCH /accounts/:id/members/:user_id を処理する。
// 既存UserのMembershipにADMIN権限を付与する(owner専用)。
func (c *MembershipController) GrantAdmin(ctx *gin.Context) {
	targetAccountID, ok := parseTargetAccountID(ctx)
	if !ok {
		return
	}

	targetUserID, err := uuid.Parse(ctx.Param("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	_, actingUserID, ok := resolveIdentity(ctx, c.accountUsecase)
	if !ok {
		return
	}

	var req grantAdminRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	view, err := c.usecase.GrantAdmin(ctx.Request.Context(), usecase.GrantAdminInput{
		ActingUserID:    actingUserID,
		TargetAccountID: targetAccountID,
		TargetUserID:    targetUserID,
	})
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, view)
}

// RemoveMember はDELETE /accounts/:id/members/:user_id を処理する。
// ownerが実行した場合は即座にMembershipが除外される。adminが実行した場合は
// ownerの承認待ちのMembershipRemovalRequestが作成される。
func (c *MembershipController) RemoveMember(ctx *gin.Context) {
	targetAccountID, ok := parseTargetAccountID(ctx)
	if !ok {
		return
	}

	targetUserID, err := uuid.Parse(ctx.Param("user_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	_, actingUserID, ok := resolveIdentity(ctx, c.accountUsecase)
	if !ok {
		return
	}

	result, err := c.usecase.RemoveUser(ctx.Request.Context(), usecase.RemoveUserInput{
		ActingUserID:    actingUserID,
		TargetAccountID: targetAccountID,
		TargetUserID:    targetUserID,
	})
	if err != nil {
		writeError(ctx, err)
		return
	}

	if result.Membership != nil {
		ctx.JSON(http.StatusOK, gin.H{"membership": result.Membership})
		return
	}

	ctx.JSON(http.StatusAccepted, gin.H{"removal_request": result.Request})
}

// ListMembers はGET /accounts/:id/members を処理する。
// Organizationに所属するMembershipを一覧取得する(owner・admin専用)。
func (c *MembershipController) ListMembers(ctx *gin.Context) {
	targetAccountID, ok := parseTargetAccountID(ctx)
	if !ok {
		return
	}

	_, actingUserID, ok := resolveIdentity(ctx, c.accountUsecase)
	if !ok {
		return
	}

	views, err := c.usecase.ListMembers(ctx.Request.Context(), actingUserID, targetAccountID)
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, views)
}

// ListRemovalRequests はGET /accounts/:id/removal-requests を処理する。
// 承認待ちのMembershipRemovalRequestを一覧取得する(owner専用)。
func (c *MembershipController) ListRemovalRequests(ctx *gin.Context) {
	targetAccountID, ok := parseTargetAccountID(ctx)
	if !ok {
		return
	}

	_, actingUserID, ok := resolveIdentity(ctx, c.accountUsecase)
	if !ok {
		return
	}

	views, err := c.usecase.ListPendingRemovalRequests(ctx.Request.Context(), actingUserID, targetAccountID)
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, views)
}

func parseRequestID(ctx *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(ctx.Param("request_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request id"})
		return uuid.Nil(), false
	}
	return id, true
}

// ApproveRemovalRequest はPOST /accounts/:id/removal-requests/:request_id/approve
// を処理する。adminが作成した除外申請をownerが承認する(owner専用)。
func (c *MembershipController) ApproveRemovalRequest(ctx *gin.Context) {
	targetAccountID, ok := parseTargetAccountID(ctx)
	if !ok {
		return
	}
	requestID, ok := parseRequestID(ctx)
	if !ok {
		return
	}

	_, actingUserID, ok := resolveIdentity(ctx, c.accountUsecase)
	if !ok {
		return
	}

	view, err := c.usecase.ApproveRemovalRequest(ctx.Request.Context(), usecase.ResolveRemovalRequestInput{
		ActingUserID:    actingUserID,
		TargetAccountID: targetAccountID,
		RequestID:       requestID,
	})
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, view)
}

// RejectRemovalRequest はPOST /accounts/:id/removal-requests/:request_id/reject
// を処理する。adminが作成した除外申請をownerが却下する(owner専用)。
func (c *MembershipController) RejectRemovalRequest(ctx *gin.Context) {
	targetAccountID, ok := parseTargetAccountID(ctx)
	if !ok {
		return
	}
	requestID, ok := parseRequestID(ctx)
	if !ok {
		return
	}

	_, actingUserID, ok := resolveIdentity(ctx, c.accountUsecase)
	if !ok {
		return
	}

	view, err := c.usecase.RejectRemovalRequest(ctx.Request.Context(), usecase.ResolveRemovalRequestInput{
		ActingUserID:    actingUserID,
		TargetAccountID: targetAccountID,
		RequestID:       requestID,
	})
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, view)
}
