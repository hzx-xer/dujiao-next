package middleware

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dujiao-next/internal/constants"
	settingsapp "github.com/dujiao-next/internal/modules/settings/application"
	"github.com/dujiao-next/internal/shared/jsonmap"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func TestKeyByIPAndJSONField(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/auth", strings.NewReader(`{"email":" Test@Example.com "}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.RemoteAddr = "1.2.3.4:5678"

	key := KeyByIPAndJSONField("email")(c)
	if key != "test@example.com|1.2.3.4" {
		t.Fatalf("key want test@example.com|1.2.3.4 got %s", key)
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		t.Fatalf("read body after key extraction failed: %v", err)
	}
	if !strings.Contains(string(body), "Test@Example.com") {
		t.Fatalf("request body should be restored after reading field")
	}
}

func TestRateLimitMiddlewareWithoutClient(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(RateLimitMiddleware(nil, RateLimitRule{WindowSeconds: 60, MaxRequests: 1}, KeyByIP))
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status want 200 got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"ok":true`) {
		t.Fatalf("expected handler response body, got %s", w.Body.String())
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/ping", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("second request must be limited even without Redis: status want 429 got %d", w.Code)
	}
}

func TestLocalRateLimiterCapacityDoesNotEvictLiveEntries(t *testing.T) {
	now := time.Now()
	limiter := &localRateLimiter{entries: make(map[string]localRateLimitEntry, localRateLimitMaxEntries)}
	for i := 0; i < localRateLimitMaxEntries; i++ {
		limiter.entries[fmt.Sprintf("live-%d", i)] = localRateLimitEntry{
			count:     1,
			expiresAt: now.Add(10 * time.Minute),
		}
	}

	rule := RateLimitRule{WindowSeconds: 60, MaxRequests: 5}
	count, _, warned := limiter.increment("new-key", rule, now)
	if count <= int64(rule.MaxRequests) {
		t.Fatalf("new key must fail closed when local limiter is at capacity, count=%d", count)
	}
	if !warned {
		t.Fatal("first capacity rejection must request an operator warning")
	}
	_, _, warned = limiter.increment("another-new-key", rule, now.Add(time.Second))
	if warned {
		t.Fatal("capacity warning must be throttled")
	}
	_, _, warned = limiter.increment("later-new-key", rule, now.Add(localRateLimitWarningInterval))
	if !warned {
		t.Fatal("capacity warning should be emitted again after the throttle interval")
	}
	if len(limiter.entries) != localRateLimitMaxEntries {
		t.Fatalf("live entry count changed: got %d want %d", len(limiter.entries), localRateLimitMaxEntries)
	}
	if _, ok := limiter.entries["live-0"]; !ok {
		t.Fatal("capacity handling must not evict live counters")
	}
	if _, ok := limiter.entries["new-key"]; ok {
		t.Fatal("rejected new key must not be stored")
	}
}

func TestLocalRateLimiterPurgesExpiredEntriesBeforeRejectingNewKey(t *testing.T) {
	now := time.Now()
	limiter := &localRateLimiter{entries: make(map[string]localRateLimitEntry, localRateLimitMaxEntries)}
	for i := 0; i < localRateLimitMaxEntries; i++ {
		expiresAt := now.Add(time.Minute)
		if i == 0 {
			expiresAt = now.Add(-time.Second)
		}
		limiter.entries[fmt.Sprintf("entry-%d", i)] = localRateLimitEntry{
			count:     1,
			expiresAt: expiresAt,
		}
	}

	count, _, warned := limiter.increment("new-key", RateLimitRule{WindowSeconds: 60, MaxRequests: 5}, now)
	if count != 1 {
		t.Fatalf("new key count = %d, want 1 after expired entry cleanup", count)
	}
	if warned {
		t.Fatal("successful expired-entry cleanup must not emit a capacity warning")
	}
	if _, ok := limiter.entries["new-key"]; !ok {
		t.Fatal("new key should be stored after expired entry cleanup")
	}
	if _, ok := limiter.entries["entry-0"]; ok {
		t.Fatal("expired entry should be removed")
	}
}

func TestRateLimitMiddlewareFallsBackLocallyWhenRedisIsUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve local address: %v", err)
	}
	addr := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatalf("close reserved address: %v", err)
	}
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		DialTimeout:  50 * time.Millisecond,
		ReadTimeout:  50 * time.Millisecond,
		WriteTimeout: 50 * time.Millisecond,
		MaxRetries:   -1,
	})
	t.Cleanup(func() { _ = client.Close() })

	r := gin.New()
	r.Use(RateLimitMiddleware(client, RateLimitRule{
		Prefix:        "redis-fallback-test",
		WindowSeconds: 60,
		MaxRequests:   1,
	}, KeyByIP))
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	first := httptest.NewRecorder()
	r.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/ping", nil))
	if first.Code != http.StatusOK {
		t.Fatalf("first fallback request status = %d, want 200", first.Code)
	}

	second := httptest.NewRecorder()
	r.ServeHTTP(second, httptest.NewRequest(http.MethodGet, "/ping", nil))
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second fallback request status = %d, want 429", second.Code)
	}
}

func TestToInt64(t *testing.T) {
	cases := []struct {
		name  string
		input interface{}
		want  int64
		ok    bool
	}{
		{name: "int64", input: int64(10), want: 10, ok: true},
		{name: "int", input: int(11), want: 11, ok: true},
		{name: "uint8", input: uint8(12), want: 12, ok: true},
		{name: "float64", input: float64(13.9), want: 13, ok: true},
		{name: "string", input: "bad", want: 0, ok: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := toInt64(tc.input)
			if ok != tc.ok {
				t.Fatalf("ok want %v got %v", tc.ok, ok)
			}
			if got != tc.want {
				t.Fatalf("value want %d got %d", tc.want, got)
			}
		})
	}
}

type callbackRoutesStoreStub struct {
	value jsonmap.JSON
}

func (s *callbackRoutesStoreStub) GetByKey(key string) (jsonmap.JSON, bool, error) {
	if key != constants.SettingKeyCallbackRoutesConfig || s.value == nil {
		return nil, false, nil
	}
	return s.value, true, nil
}

func (s *callbackRoutesStoreStub) Upsert(_ string, value jsonmap.JSON) (jsonmap.JSON, error) {
	return value, nil
}

// 自定义回调路径在全局中间件里直接分发，绕开了带限流的回调路由组；
// CustomCallbackRateLimitMiddleware 必须给它们套上同样的限流，其他请求不受影响。
func TestCustomCallbackRateLimitAppliesOnlyToCustomCallbackPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := settingsapp.NewService(&callbackRoutesStoreStub{value: jsonmap.JSON{
		constants.SettingFieldPaymentCallback: "/api/v1/pay-notify-x7",
		constants.SettingFieldStripeWebhook:   "/api/v1/stripe-hook-x7",
	}})
	service.InvalidateCallbackRoutesCache()
	t.Cleanup(service.InvalidateCallbackRoutesCache)

	rule := RateLimitRule{Prefix: "test:custom_callback", WindowSeconds: 60, MaxRequests: 2, BlockSeconds: 60}
	r := gin.New()
	r.Use(CustomCallbackRateLimitMiddleware(service, RateLimitMiddleware(nil, rule, KeyByIP)))
	reached := 0
	r.Use(func(c *gin.Context) {
		reached++
		c.Status(http.StatusOK)
		c.Abort()
	})

	do := func(method, path string) int {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(method, path, nil)
		req.RemoteAddr = "203.0.113.7:1234"
		r.ServeHTTP(w, req)
		return w.Code
	}

	for i := 0; i < 2; i++ {
		if code := do(http.MethodPost, "/api/v1/pay-notify-x7"); code != http.StatusOK {
			t.Fatalf("request %d within limit: status %d", i+1, code)
		}
	}
	if code := do(http.MethodPost, "/api/v1/pay-notify-x7"); code != http.StatusTooManyRequests {
		t.Fatalf("custom payment callback over limit: status %d, want 429", code)
	}
	// 同一来源 IP 的计数在所有自定义回调路径间共享，与默认回调路由组一致
	if code := do(http.MethodPost, "/api/v1/stripe-hook-x7"); code != http.StatusTooManyRequests {
		t.Fatalf("custom webhook over shared limit: status %d, want 429", code)
	}
	// 非回调路径、GET 访问只接受 POST 的 webhook 路径：不受这层限流影响
	for _, tc := range []struct{ method, path string }{
		{http.MethodPost, "/api/v1/public/config"},
		{http.MethodGet, "/api/v1/stripe-hook-x7"},
		{http.MethodGet, "/health"},
	} {
		if code := do(tc.method, tc.path); code != http.StatusOK {
			t.Fatalf("%s %s should pass through: status %d", tc.method, tc.path, code)
		}
	}
	if reached != 5 {
		t.Fatalf("downstream handler reached %d times, want 5", reached)
	}
}

func TestCustomCallbackRateLimitPassesThroughWithoutCustomRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := settingsapp.NewService(&callbackRoutesStoreStub{})
	service.InvalidateCallbackRoutesCache()
	t.Cleanup(service.InvalidateCallbackRoutesCache)

	rule := RateLimitRule{Prefix: "test:custom_callback_none", WindowSeconds: 60, MaxRequests: 1, BlockSeconds: 60}
	r := gin.New()
	r.Use(CustomCallbackRateLimitMiddleware(service, RateLimitMiddleware(nil, rule, KeyByIP)))
	r.POST("/api/v1/payments/callback", func(c *gin.Context) { c.Status(http.StatusOK) })

	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/payments/callback", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: status %d, want 200 (no custom routes configured)", i+1, w.Code)
		}
	}
}
