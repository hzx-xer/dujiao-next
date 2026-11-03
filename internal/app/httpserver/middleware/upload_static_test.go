package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

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
