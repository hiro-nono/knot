// Package router はAPIエンドポイントをcontrollerに紐付ける。
package router

import (
	"github.com/gin-gonic/gin"

	"knot-api/internal/controller"
)

// New はginエンジンを生成し、エンドポイントをcontrollerに紐付ける。
//
// authMiddlewareはJWT検証を必須とするginミドルウェアで、
// optionalAuthMiddlewareはJWTがあれば検証してcontextに載せるが無くても
// リクエストを中断しないミドルウェア。corsMiddlewareは許可されたオリジンからの
// クロスオリジンリクエストのみを許可し、csrfMiddlewareは状態変更リクエストに
// ついてCSRFトークン(二重提出Cookie方式)を検証する。csrfIssueHandlerは
// そのCSRFトークンを新規発行するハンドラ。いずれもcomposition root
// (cmd/server/main.go)が生成して渡す。routerはその中身(infrastructure)を
// 知らず、gin.HandlerFuncとしてのみ扱う。
func New(
	health *controller.HealthController,
	information *controller.InformationController,
	display *controller.DisplayController,
	response *controller.ResponseController,
	account *controller.AccountController,
	membership *controller.MembershipController,
	authMiddleware gin.HandlerFunc,
	optionalAuthMiddleware gin.HandlerFunc,
	corsMiddleware gin.HandlerFunc,
	csrfMiddleware gin.HandlerFunc,
	csrfIssueHandler gin.HandlerFunc,
) *gin.Engine {
	r := gin.Default()
	r.Use(corsMiddleware)
	r.Use(csrfMiddleware)

	r.GET("/health", health.Health)
	r.GET("/csrf-token", csrfIssueHandler)

	informations := r.Group("/informations", authMiddleware)
	informations.GET("", information.ListMine)
	informations.POST("/messages", information.ProcessInformation)
	informations.POST("/:id/recipients", information.AddRecipient)
	informations.GET("/:id/recipients", information.ListRecipients)
	informations.DELETE("/:id/recipients/:user_id", information.RemoveRecipient)
	informations.POST("/:id/display", display.GenerateDisplay)
	informations.POST("/:id/display/comparisons", display.PrepareComparison)
	informations.POST("/:id/display/comparisons/selection", display.SelectComparison)
	informations.POST("/:id/display/chat", display.Chat)
	informations.GET("/:id/responses", response.ListResponses)

	// PUBLIC+ANONYMOUSなInformationは未ログインでも回答・質問一覧の閲覧ができるため、
	// この2つのエンドポイントのみ認証を必須としない(認証情報はあれば利用する)。
	r.GET("/informations/:id/sources", optionalAuthMiddleware, display.ListSources)
	r.POST("/informations/:id/responses", optionalAuthMiddleware, response.SubmitResponse)

	accounts := r.Group("/accounts", authMiddleware)
	accounts.POST("", account.Register)
	accounts.GET("", account.ListByStatus)
	accounts.GET("/me", account.GetMe)
	accounts.PATCH("/me", account.UpdateMyProfile)
	accounts.DELETE("/me", account.Delete)
	accounts.PATCH("/:id/status", account.UpdateStatus)
	accounts.POST("/:id/members", membership.AddMember)
	accounts.GET("/:id/members", membership.ListMembers)
	accounts.PATCH("/:id/members/:user_id", membership.GrantAdmin)
	accounts.DELETE("/:id/members/:user_id", membership.RemoveMember)
	accounts.GET("/:id/removal-requests", membership.ListRemovalRequests)
	accounts.POST("/:id/removal-requests/:request_id/approve", membership.ApproveRemovalRequest)
	accounts.POST("/:id/removal-requests/:request_id/reject", membership.RejectRemovalRequest)

	r.GET("/users/:id", authMiddleware, account.GetUserProfile)

	return r
}
