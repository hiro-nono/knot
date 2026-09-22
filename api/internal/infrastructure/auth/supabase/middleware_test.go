package supabase_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"knot-api/internal/infrastructure/auth/supabase"
	"knot-api/internal/infrastructure/auth/supabase/supabasetest"
)

func newTestMiddlewareRouter(verifier *supabase.Verifier) *gin.Engine {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/protected", supabase.Middleware(verifier), func(ctx *gin.Context) {
		claims, ok := supabase.FromContext(ctx)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "no claims"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"user_id": claims.UserID})
	})
	return r
}

func TestMiddleware_MissingHeader(t *testing.T) {
	keySet := supabasetest.NewKeySet()
	t.Cleanup(keySet.Close)
	r := newTestMiddlewareRouter(keySet.NewVerifier(context.Background()))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestMiddleware_InvalidToken(t *testing.T) {
	keySet := supabasetest.NewKeySet()
	t.Cleanup(keySet.Close)
	r := newTestMiddlewareRouter(keySet.NewVerifier(context.Background()))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer not-a-jwt")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestMiddleware_ValidToken(t *testing.T) {
	keySet := supabasetest.NewKeySet()
	t.Cleanup(keySet.Close)
	r := newTestMiddlewareRouter(keySet.NewVerifier(context.Background()))

	signed := keySet.SignToken("user-1")

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func newTestOptionalMiddlewareRouter(verifier *supabase.Verifier) *gin.Engine {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/optional", supabase.OptionalMiddleware(verifier), func(ctx *gin.Context) {
		claims, ok := supabase.FromContext(ctx)
		if !ok {
			ctx.JSON(http.StatusOK, gin.H{"authenticated": false})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"authenticated": true, "user_id": claims.UserID})
	})
	return r
}

func TestOptionalMiddleware_MissingHeader_ProceedsUnauthenticated(t *testing.T) {
	keySet := supabasetest.NewKeySet()
	t.Cleanup(keySet.Close)
	r := newTestOptionalMiddlewareRouter(keySet.NewVerifier(context.Background()))

	req := httptest.NewRequest(http.MethodGet, "/optional", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"authenticated":false`) {
		t.Errorf("body = %s, want authenticated:false", w.Body.String())
	}
}

func TestOptionalMiddleware_InvalidToken_ProceedsUnauthenticated(t *testing.T) {
	keySet := supabasetest.NewKeySet()
	t.Cleanup(keySet.Close)
	r := newTestOptionalMiddlewareRouter(keySet.NewVerifier(context.Background()))

	req := httptest.NewRequest(http.MethodGet, "/optional", nil)
	req.Header.Set("Authorization", "Bearer not-a-jwt")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"authenticated":false`) {
		t.Errorf("body = %s, want authenticated:false", w.Body.String())
	}
}

func TestOptionalMiddleware_ValidToken_SetsClaims(t *testing.T) {
	keySet := supabasetest.NewKeySet()
	t.Cleanup(keySet.Close)
	r := newTestOptionalMiddlewareRouter(keySet.NewVerifier(context.Background()))

	signed := keySet.SignToken("user-1")

	req := httptest.NewRequest(http.MethodGet, "/optional", nil)
	req.Header.Set("Authorization", "Bearer "+signed)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"authenticated":true`) {
		t.Errorf("body = %s, want authenticated:true", w.Body.String())
	}
}
