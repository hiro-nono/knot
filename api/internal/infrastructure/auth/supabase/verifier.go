package supabase

import (
	"context"
	"errors"
	"fmt"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

// jwksPath はSupabaseプロジェクトが公開する公開鍵セット(JWKS)のパス。
const jwksPath = "/auth/v1/.well-known/jwks.json"

// allowedSigningMethods はVerifyが受理する署名アルゴリズム。
// Supabaseの非対称鍵署名(ES256)のみを許可し、HS256(共有シークレット)は
// 受理しない。allowlistを明示することで、alg混同攻撃(JWKSの公開鍵を
// HMACシークレットとして悪用する攻撃)を防ぐ。
var allowedSigningMethods = []string{"ES256", "RS256"}

// Verifier はSupabaseが発行したJWTを検証する。
//
// SupabaseプロジェクトのJWKS(公開鍵セット)を使って非対称鍵署名を検証する。
// 秘密鍵はSupabase側にのみ存在し、このアプリケーションは公開鍵しか扱わない。
type Verifier struct {
	keyfunc keyfunc.Keyfunc
}

// NewVerifier はVerifierを生成する。supabaseURLはSupabaseプロジェクトの
// ベースURL(例: https://xxxx.supabase.co)。JWKSはバックグラウンドで定期的に
// 再取得され、Supabase側の鍵ローテーションに追従する。
//
// ctxはJWKSの定期再取得を行うバックグラウンドgoroutineの生存期間を制御する。
// プロセスの生存期間と一致させ、シャットダウン時にキャンセルすること。
func NewVerifier(ctx context.Context, supabaseURL string) (*Verifier, error) {
	kf, err := keyfunc.NewDefaultCtx(ctx, []string{supabaseURL + jwksPath})
	if err != nil {
		return nil, fmt.Errorf("create JWKS keyfunc: %w", err)
	}
	return &Verifier{keyfunc: kf}, nil
}

// supabaseClaims はSupabaseが発行するJWTのペイロードのうち、
// このアプリケーションが利用する部分を表す。
type supabaseClaims struct {
	jwt.RegisteredClaims
}

// Verify はtokenStringの署名・有効期限を検証し、Claimsを抽出する。
func (v *Verifier) Verify(tokenString string) (*Claims, error) {
	var claims supabaseClaims
	token, err := jwt.ParseWithClaims(
		tokenString,
		&claims,
		v.keyfunc.Keyfunc,
		jwt.WithValidMethods(allowedSigningMethods),
	)
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	if claims.Subject == "" {
		return nil, errors.New("token has no subject")
	}

	return &Claims{
		UserID: claims.Subject,
	}, nil
}
