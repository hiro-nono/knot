package csrf

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newTestRouter(cfg Config) *gin.Engine {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/csrf-token", IssueTokenHandler(cfg))
	protected := r.Group("/protected", Middleware(cfg))
	protected.POST("", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

func TestIssueTokenHandler_SetsCookieAndBody(t *testing.T) {
	r := newTestRouter(Config{Secure: false})

	req := httptest.NewRequest(http.MethodGet, "/csrf-token", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	cookies := w.Result().Cookies()
	var found *http.Cookie
	for _, c := range cookies {
		if c.Name == CookieName {
			found = c
			break
		}
	}
	if found == nil {
		t.Fatalf("cookie %q not set", CookieName)
	}
	if found.Value == "" {
		t.Error("cookie value is empty")
	}
	if w.Body.Len() == 0 {
		t.Error("response body is empty")
	}
}

func TestMiddleware_GET_NotChecked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Middleware(Config{Secure: false}))
	r.GET("/resource", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/resource", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestMiddleware_POST_MissingCookie_Forbidden(t *testing.T) {
	r := newTestRouter(Config{Secure: false})

	req := httptest.NewRequest(http.MethodPost, "/protected", nil)
	req.Header.Set(HeaderName, "some-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestMiddleware_POST_MissingHeader_Forbidden(t *testing.T) {
	r := newTestRouter(Config{Secure: false})

	req := httptest.NewRequest(http.MethodPost, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: CookieName, Value: "some-token"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestMiddleware_POST_MismatchedToken_Forbidden(t *testing.T) {
	r := newTestRouter(Config{Secure: false})

	req := httptest.NewRequest(http.MethodPost, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: CookieName, Value: "cookie-token"})
	req.Header.Set(HeaderName, "different-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestMiddleware_POST_MatchingToken_OK(t *testing.T) {
	r := newTestRouter(Config{Secure: false})

	req := httptest.NewRequest(http.MethodPost, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: CookieName, Value: "matching-token"})
	req.Header.Set(HeaderName, "matching-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestGenerateToken_ProducesUniqueValues(t *testing.T) {
	a, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	b, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	if a == b {
		t.Error("GenerateToken produced identical tokens twice")
	}
	if a == "" || b == "" {
		t.Error("GenerateToken produced empty token")
	}
}
