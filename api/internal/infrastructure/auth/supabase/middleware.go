package supabase

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const claimsContextKey = "supabase_claims"
const bearerPrefix = "Bearer "

// Middleware はAuthorizationヘッダのBearerトークンをVerifierで検証し、
// 検証済みのClaims(UserID)をginのcontextに設定するミドルウェアを返す。
// トークンが無い、または検証に失敗した場合は401を返しリクエストを中断する。
func Middleware(verifier *Verifier) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		header := ctx.GetHeader("Authorization")
		if !strings.HasPrefix(header, bearerPrefix) {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

		token := strings.TrimPrefix(header, bearerPrefix)
		claims, err := verifier.Verify(token)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		ctx.Set(claimsContextKey, claims)
		ctx.Next()
	}
}

// OptionalMiddleware はAuthorizationヘッダのBearerトークンを検証できた場合のみ
// Claimsをginのcontextに設定するミドルウェアを返す。トークンが無い、または
// 検証に失敗した場合でもリクエストは中断せず、未認証のまま処理を続行する。
//
// PUBLIC+ANONYMOUSなInformationへの回答のように、未ログインでもアクセスを
// 許可しつつ、ログイン済みであればその情報も使いたいエンドポイント専用。
// 認証が必須なエンドポイントには引き続きMiddlewareを使うこと。
func OptionalMiddleware(verifier *Verifier) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		header := ctx.GetHeader("Authorization")
		if !strings.HasPrefix(header, bearerPrefix) {
			ctx.Next()
			return
		}

		token := strings.TrimPrefix(header, bearerPrefix)
		claims, err := verifier.Verify(token)
		if err != nil {
			ctx.Next()
			return
		}

		ctx.Set(claimsContextKey, claims)
		ctx.Next()
	}
}

// FromContext はMiddlewareが設定したClaimsを取得する。
// Middlewareが適用されていないルートで呼び出した場合はok=falseを返す。
func FromContext(ctx *gin.Context) (*Claims, bool) {
	value, ok := ctx.Get(claimsContextKey)
	if !ok {
		return nil, false
	}
	claims, ok := value.(*Claims)
	return claims, ok
}
