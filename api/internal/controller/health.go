package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthController はサービスの生存状態を報告する。
type HealthController struct{}

// NewHealthController はHealthControllerを生成する。
func NewHealthController() *HealthController {
	return &HealthController{}
}

// Health は単純な生存確認用のレスポンスを返す。
func (c *HealthController) Health(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}
