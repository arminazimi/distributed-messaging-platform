package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

func TestGlobalRateLimitAllowsNormalRequest(t *testing.T) {
	Configure(10, 10, 10, 10)

	e := echo.New()
	e.Use(GlobalMiddleware())
	e.GET("/ok", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestGlobalRateLimitRejectsWhenExceeded(t *testing.T) {
	Configure(1, 1, 10, 10)

	e := echo.New()
	e.Use(GlobalMiddleware())
	e.GET("/ok", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/ok", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if i == 0 && rec.Code != http.StatusOK {
			t.Fatalf("expected first request to pass, got %d", rec.Code)
		}
		if i == 1 && rec.Code != http.StatusTooManyRequests {
			t.Fatalf("expected second request to be rate limited, got %d", rec.Code)
		}
	}
}

func TestUserRateLimitRejectsOnlySameUser(t *testing.T) {
	limiter := NewUserLimiter(1, 1, time.Minute)

	if !limiter.Allow("1") {
		t.Fatalf("expected first request for user 1 to pass")
	}
	if limiter.Allow("1") {
		t.Fatalf("expected second request for user 1 to be rate limited")
	}
	if !limiter.Allow("2") {
		t.Fatalf("expected first request for user 2 to pass")
	}
}
