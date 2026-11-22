package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dujiao-next/internal/config"

	"github.com/gin-gonic/gin"
)

func TestResolveAllowedOrigin(t *testing.T) {
	got := resolveAllowedOrigin("https://example.com", []string{"*"})
	if got != "*" {
		t.Fatalf("wildcard should return *, got %s", got)
	}

	got = resolveAllowedOrigin("https://a.example.com", []string{"https://a.example.com", "https://b.example.com"})
	if got != "https://a.example.com" {
		t.Fatalf("allow-list should return matched origin, got %s", got)
	}

	got = resolveAllowedOrigin("https://x.example.com", []string{"https://a.example.com"})
	if got != "" {
		t.Fatalf("unmatched origin should be empty, got %s", got)
	}
}

func TestCORSMiddlewareCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const origin = "https://evil.example"
	tests := []struct {
		name            string
		allowed         []string
		credentials     bool
		wantOrigin      string
		wantCredentials string
	}{
		{name: "wildcard with credentials never echoes the origin", allowed: []string{"*"}, credentials: true, wantOrigin: "*"},
		{name: "wildcard without credentials", allowed: []string{"*"}, wantOrigin: "*"},
		{name: "allow-listed origin can use credentials", allowed: []string{origin}, credentials: true, wantOrigin: origin, wantCredentials: "true"},
		{name: "allow-listed origin without credentials", allowed: []string{origin}, wantOrigin: origin},
		{name: "unlisted origin gets nothing", allowed: []string{"https://shop.example"}, credentials: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(CORSMiddleware(config.CORSConfig{AllowedOrigins: tt.allowed, AllowCredentials: tt.credentials}))
			r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

			for _, method := range []string{http.MethodGet, http.MethodOptions} {
				req := httptest.NewRequest(method, "/ping", nil)
				req.Header.Set("Origin", origin)
				rec := httptest.NewRecorder()
				r.ServeHTTP(rec, req)
				if got := rec.Header().Get("Access-Control-Allow-Origin"); got != tt.wantOrigin {
					t.Fatalf("%s Access-Control-Allow-Origin = %q, want %q", method, got, tt.wantOrigin)
				}
				if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != tt.wantCredentials {
					t.Fatalf("%s Access-Control-Allow-Credentials = %q, want %q", method, got, tt.wantCredentials)
				}
			}
		})
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(RequestIDMiddleware())
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"request_id": getRequestID(c)})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set(requestIDHeader, "req-123")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status want 200 got %d", w.Code)
	}
	if w.Header().Get(requestIDHeader) != "req-123" {
		t.Fatalf("response request id want req-123 got %s", w.Header().Get(requestIDHeader))
	}
	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp["request_id"] != "req-123" {
		t.Fatalf("context request id want req-123 got %s", resp["request_id"])
	}

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/ping", nil)
	r.ServeHTTP(w2, req2)
	generated := w2.Header().Get(requestIDHeader)
	if generated == "" {
		t.Fatalf("generated request id should not be empty")
	}
	if resp := strings.TrimSpace(generated); resp == "" {
		t.Fatalf("generated request id should not be blank")
	}
}

func TestJWTAuthMiddlewareMissingSecret(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(JWTAuthMiddleware("", nil))
	r.GET("/admin/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/ping", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status want 200 got %d", w.Code)
	}
	var resp struct {
		StatusCode int `json:"status_code"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.StatusCode != 401 {
		t.Fatalf("status_code want 401 got %d", resp.StatusCode)
	}
}

func newUploadStaticTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	dir := t.TempDir()
	for _, name := range []string{"a.png", "b.svg"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("content"), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Group("/uploads", UploadStatic()).Static("/", dir)
	return r
}

func serveUpload(r *gin.Engine, target string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestUploadStaticCachesExistingFiles(t *testing.T) {
	rec := serveUpload(newUploadStaticTestRouter(t), "/uploads/a.png", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("Cache-Control"); got != uploadCacheControl {
		t.Fatalf("Cache-Control = %q, want %q", got, uploadCacheControl)
	}
}

func TestUploadStaticKeepsCacheHeaderOnNotModified(t *testing.T) {
	r := newUploadStaticTestRouter(t)
	first := serveUpload(r, "/uploads/a.png", nil)
	rec := serveUpload(r, "/uploads/a.png", map[string]string{"If-Modified-Since": first.Header().Get("Last-Modified")})
	if rec.Code != http.StatusNotModified {
		t.Fatalf("status = %d, want 304", rec.Code)
	}
	if got := rec.Header().Get("Cache-Control"); got != uploadCacheControl {
		t.Fatalf("Cache-Control = %q, want %q", got, uploadCacheControl)
	}
}

// 不存在的文件不能被缓存，否则图片补传后浏览器和 CDN 仍会长期返回 404。
func TestUploadStaticDoesNotCacheMissingFiles(t *testing.T) {
	rec := serveUpload(newUploadStaticTestRouter(t), "/uploads/missing.png", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	if got := rec.Header().Get("Cache-Control"); got != "" {
		t.Fatalf("Cache-Control = %q, want empty on 404", got)
	}
}

func TestUploadStaticKeepsSecurityHeaders(t *testing.T) {
	r := newUploadStaticTestRouter(t)
	for _, target := range []string{"/uploads/a.png", "/uploads/missing.png"} {
		rec := serveUpload(r, target, nil)
		if got := rec.Header().Get("Content-Security-Policy"); got != "sandbox; script-src 'none'" {
			t.Fatalf("%s: Content-Security-Policy = %q", target, got)
		}
		if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
			t.Fatalf("%s: X-Content-Type-Options = %q", target, got)
		}
	}
	if got := serveUpload(r, "/uploads/b.svg", nil).Header().Get("Content-Disposition"); got != "attachment" {
		t.Fatalf("svg Content-Disposition = %q, want attachment", got)
	}
	if got := serveUpload(r, "/uploads/a.png", nil).Header().Get("Content-Disposition"); got != "" {
		t.Fatalf("png Content-Disposition = %q, want empty", got)
	}
}
