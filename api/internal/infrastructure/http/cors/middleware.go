// Package cors はCross-Origin Resource Sharing(CORS)を扱うginミドルウェアを提供する。
package cors

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Config はCORSミドルウェアの設定を表す。
type Config struct {
	// AllowedOrigins はcrossOriginリクエストを許可するオリジンの一覧。
	AllowedOrigins []string
}

// Middleware は設定されたオリジンからのクロスオリジンリクエストのみを許可する
// ginミドルウェアを返す。CSRFトークンをCookieでやり取りするため
// Access-Control-Allow-Credentialsを付与する必要があり、その場合
// Access-Control-Allow-Originに"*"は使えないため、許可リストに含まれる
// オリジンのみをそのまま反映する。
func Middleware(cfg Config) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(cfg.AllowedOrigins))
	for _, origin := range cfg.AllowedOrigins {
		allowed[origin] = struct{}{}
	}

	return func(ctx *gin.Context) {
		origin := ctx.GetHeader("Origin")
		if origin != "" {
			ctx.Header("Vary", "Origin")
			if _, ok := allowed[origin]; ok {
				ctx.Header("Access-Control-Allow-Origin", origin)
				ctx.Header("Access-Control-Allow-Credentials", "true")
			}
		}

		if ctx.Request.Method == http.MethodOptions {
			ctx.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			ctx.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")
			ctx.Header("Access-Control-Max-Age", "600")
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		ctx.Next()
	}
}
