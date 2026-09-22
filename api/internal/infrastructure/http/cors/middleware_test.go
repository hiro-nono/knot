package cors

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newTestRouter(allowedOrigins []string) *gin.Engine {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(Middleware(Config{AllowedOrigins: allowedOrigins}))
	r.GET("/resource", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

func TestMiddleware_AllowedOrigin_SetsHeaders(t *testing.T) {
	r := newTestRouter([]string{"http://localhost:3000"})

	req := httptest.NewRequest(http.MethodGet, "/resource", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "http://localhost:3000")
	}
	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("Access-Control-Allow-Credentials = %q, want %q", got, "true")
	}
	if got := w.Header().Get("Vary"); got != "Origin" {
		t.Errorf("Vary = %q, want %q", got, "Origin")
	}
}

func TestMiddleware_DisallowedOrigin_OmitsHeaders(t *testing.T) {
	r := newTestRouter([]string{"http://localhost:3000"})

	req := httptest.NewRequest(http.MethodGet, "/resource", nil)
	req.Header.Set("Origin", "http://evil.example.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Access-Control-Allow-Origin = %q, want empty", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Errorf("Access-Control-Allow-Credentials = %q, want empty", got)
	}
}

func TestMiddleware_NoOrigin_ProceedsWithoutCORSHeaders(t *testing.T) {
	r := newTestRouter([]string{"http://localhost:3000"})

	req := httptest.NewRequest(http.MethodGet, "/resource", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Access-Control-Allow-Origin = %q, want empty", got)
	}
}

func TestMiddleware_Preflight_RespondsNoContent(t *testing.T) {
	r := newTestRouter([]string{"http://localhost:3000"})

	req := httptest.NewRequest(http.MethodOptions, "/resource", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
	if got := w.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("Access-Control-Allow-Methods = empty, want non-empty")
	}
	if got := w.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Error("Access-Control-Allow-Headers = empty, want non-empty")
	}
}
