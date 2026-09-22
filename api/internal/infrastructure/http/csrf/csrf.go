// Package csrf は二重提出Cookie(double-submit cookie)方式によるCSRF対策を提供する。
//
// 認証自体はAuthorization: BearerヘッダーのJWTで行っており、Cookieは使っていない。
// CSRFはCookieベースの認証に対して意味を持つ防御であるため、本来この構成では
// 必須ではないが、追加の防御層として発行・検証の両方を提供する。サーバー側に
// 状態を持たない(トークンをDBやメモリに保存しない)ステートレスな方式であり、
// JWT中心のこのAPIの構成と一貫性がある。
package csrf

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// CookieName はCSRFトークンを保持するCookieの名前。
	CookieName = "csrf_token"
	// HeaderName はクライアントがCSRFトークンを送り返すリクエストヘッダーの名前。
	HeaderName = "X-CSRF-Token"

	tokenBytes   = 32
	cookieMaxAge = 24 * time.Hour
)

// Config はCSRFトークン発行・検証の設定を表す。
type Config struct {
	// Secure はCookieにSecure属性(HTTPS必須)を付けるかどうか。
	// 本番(GinMode=release)ではtrue、ローカル開発(http://localhost)ではfalseにする。
	// trueの場合、フロントエンドとAPIが別オリジンのクロスサイト構成を想定し
	// SameSite=Noneを、falseの場合は同一サイト前提でSameSite=Laxを使う。
	Secure bool
}

// GenerateToken は暗号学的に安全な乱数からCSRFトークンを新規生成する。
func GenerateToken() (string, error) {
	b := make([]byte, tokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate csrf token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// IssueTokenHandler はCSRFトークンを新規発行するginハンドラを返す。
// トークンはCookie(HttpOnly=false、クライアントJSから読み取り可能)に設定すると
// 同時に、利便性のためレスポンスボディにも含める。クライアントは、以降の
// 状態変更リクエスト(POST/PUT/PATCH/DELETE)でこのトークンをX-CSRF-Token
// ヘッダーに設定する必要がある。
func IssueTokenHandler(cfg Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := GenerateToken()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		setCookie(ctx, cfg, token)
		ctx.JSON(http.StatusOK, gin.H{"csrf_token": token})
	}
}

// Middleware はPOST/PUT/PATCH/DELETEリクエストについて、X-CSRF-Tokenヘッダーと
// csrf_token Cookieの値が一致することを検証するginミドルウェアを返す
// (二重提出Cookie方式)。いずれかが無い、または値が一致しない場合は403を返す。
// GET/HEAD/OPTIONSは対象外(状態を変更しないため)。
func Middleware(cfg Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !requiresCheck(ctx.Request.Method) {
			ctx.Next()
			return
		}

		cookieToken, err := ctx.Cookie(CookieName)
		if err != nil || cookieToken == "" {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "missing csrf cookie"})
			return
		}

		headerToken := ctx.GetHeader(HeaderName)
		if headerToken == "" || headerToken != cookieToken {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid csrf token"})
			return
		}

		ctx.Next()
	}
}

func requiresCheck(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func setCookie(ctx *gin.Context, cfg Config, token string) {
	sameSite := http.SameSiteLaxMode
	if cfg.Secure {
		sameSite = http.SameSiteNoneMode
	}
	ctx.SetSameSite(sameSite)
	ctx.SetCookie(CookieName, token, int(cookieMaxAge.Seconds()), "/", "", cfg.Secure, false)
}
