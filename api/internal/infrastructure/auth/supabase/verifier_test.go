package supabase_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"knot-api/internal/infrastructure/auth/supabase"
	"knot-api/internal/infrastructure/auth/supabase/supabasetest"
)

func TestVerifier_Verify_Success(t *testing.T) {
	keySet := supabasetest.NewKeySet()
	t.Cleanup(keySet.Close)
	verifier := keySet.NewVerifier(context.Background())

	tokenString := keySet.SignToken("user-123")

	claims, err := verifier.Verify(tokenString)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("UserID = %q, want %q", claims.UserID, "user-123")
	}
}

func TestVerifier_Verify_UnknownKey(t *testing.T) {
	// トークンはJWKSに登録されていない(別プロジェクトの)鍵で署名されている。
	signingKeySet := supabasetest.NewKeySet()
	t.Cleanup(signingKeySet.Close)
	tokenString := signingKeySet.SignToken("user-123")

	verifyingKeySet := supabasetest.NewKeySet()
	t.Cleanup(verifyingKeySet.Close)
	verifier := verifyingKeySet.NewVerifier(context.Background())

	if _, err := verifier.Verify(tokenString); err == nil {
		t.Error("Verify() error = nil, want error for token signed with unknown key")
	}
}

func TestVerifier_Verify_WrongSigningMethod(t *testing.T) {
	keySet := supabasetest.NewKeySet()
	t.Cleanup(keySet.Close)
	verifier := keySet.NewVerifier(context.Background())

	// HS256(共有シークレット)で署名されたトークンは、公開鍵をシークレットとして
	// 悪用するalg混同攻撃の典型例であり、拒否されなければならない。
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "user-123",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	signed, err := token.SignedString([]byte("attacker-controlled"))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	if _, err := verifier.Verify(signed); err == nil {
		t.Error("Verify() error = nil, want error for HS256-signed token")
	}
}

func TestVerifier_Verify_Expired(t *testing.T) {
	keySet := supabasetest.NewKeySet()
	t.Cleanup(keySet.Close)
	verifier := keySet.NewVerifier(context.Background())

	tokenString := keySet.SignClaims(jwt.MapClaims{
		"sub": "user-123",
		"exp": time.Now().Add(-time.Hour).Unix(),
	})

	if _, err := verifier.Verify(tokenString); err == nil {
		t.Error("Verify() error = nil, want error for expired token")
	}
}

func TestVerifier_Verify_MissingSubject(t *testing.T) {
	keySet := supabasetest.NewKeySet()
	t.Cleanup(keySet.Close)
	verifier := keySet.NewVerifier(context.Background())

	tokenString := keySet.SignClaims(jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	if _, err := verifier.Verify(tokenString); err == nil {
		t.Error("Verify() error = nil, want error for missing subject")
	}
}

func TestVerifier_Verify_MalformedToken(t *testing.T) {
	keySet := supabasetest.NewKeySet()
	t.Cleanup(keySet.Close)
	verifier := keySet.NewVerifier(context.Background())

	if _, err := verifier.Verify("not-a-jwt"); err == nil {
		t.Error("Verify() error = nil, want error for malformed token")
	}
}

func TestNewVerifier_UnreachableJWKS(t *testing.T) {
	// keyfuncはJWKS初回取得の失敗ではエラーを返さず、バックグラウンドで
	// 再取得を試み続ける(起動時の一時的なネットワーク障害でクラッシュ
	// ループさせないため)。その代わり、鍵が1つも読み込めていない間は
	// どんなトークンの検証も失敗する。
	verifier, err := supabase.NewVerifier(context.Background(), "http://127.0.0.1:0")
	if err != nil {
		t.Fatalf("NewVerifier() error = %v, want nil", err)
	}

	keySet := supabasetest.NewKeySet()
	t.Cleanup(keySet.Close)
	tokenString := keySet.SignToken("user-123")
	if _, err := verifier.Verify(tokenString); err == nil {
		t.Error("Verify() error = nil, want error when no JWKS keys are loaded")
	}
}
