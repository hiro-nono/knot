// Package supabasetest はSupabaseのJWKS(公開鍵セット)によるJWT署名検証を
// テストするためのヘルパーを提供する。本番コードからは参照しないこと。
package supabasetest

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/MicahParks/jwkset"
	"github.com/golang-jwt/jwt/v5"

	"knot-api/internal/infrastructure/auth/supabase"
)

const testKID = "test-kid"

// KeySet はテスト用のECDSA鍵ペアと、その公開鍵をJWKSとして配信する
// httptest.Serverを保持する。Supabaseプロジェクトが非対称鍵(ES256)で
// JWTに署名する状況を模倣する。
//
// 鍵生成やJWKSの組み立ては実質的に失敗し得ないため、コンストラクタは
// *testing.Tを要求せずpanicする。パッケージレベルの共有フィクスチャとして
// `var testKeySet = supabasetest.NewKeySet()` の形で使うことを想定している。
type KeySet struct {
	Server     *httptest.Server
	PrivateKey *ecdsa.PrivateKey
}

// NewKeySet はテスト用のECDSA鍵ペアを生成し、公開鍵をJWKSとして配信する
// httptest.Serverを起動する。呼び出し側はテスト終了時にCloseを呼ぶこと
// (プロセス生存期間で使い回すパッケージレベルのフィクスチャの場合は不要)。
func NewKeySet() *KeySet {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic("supabasetest: generate key: " + err.Error())
	}

	jwk, err := jwkset.NewJWKFromKey(privateKey.Public(), jwkset.JWKOptions{
		Metadata: jwkset.JWKMetadataOptions{
			KID: testKID,
			ALG: jwkset.AlgES256,
			USE: jwkset.UseSig,
		},
	})
	if err != nil {
		panic("supabasetest: build JWK: " + err.Error())
	}

	body, err := json.Marshal(jwkset.JWKSMarshal{Keys: []jwkset.JWKMarshal{jwk.Marshal()}})
	if err != nil {
		panic("supabasetest: marshal JWKS: " + err.Error())
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))

	return &KeySet{Server: server, PrivateKey: privateKey}
}

// Close はJWKS配信サーバーを終了する。
func (ks *KeySet) Close() {
	ks.Server.Close()
}

// NewVerifier はKeySetのJWKSを参照するsupabase.Verifierを生成する。
func (ks *KeySet) NewVerifier(ctx context.Context) *supabase.Verifier {
	verifier, err := supabase.NewVerifier(ctx, ks.Server.URL)
	if err != nil {
		panic("supabasetest: NewVerifier: " + err.Error())
	}
	return verifier
}

// SignToken はKeySetの秘密鍵でuserIDをsubとするJWT(有効期限1時間後)に署名する。
func (ks *KeySet) SignToken(userID string) string {
	return ks.SignClaims(jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour).Unix(),
	})
}

// SignClaims はKeySetの秘密鍵で任意のclaimsに署名する。
// 期限切れやsub欠落など、異常系のテストで使う。
func (ks *KeySet) SignClaims(claims jwt.Claims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = testKID
	signed, err := token.SignedString(ks.PrivateKey)
	if err != nil {
		panic("supabasetest: sign token: " + err.Error())
	}
	return signed
}
