package health

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gateway/app"

	"github.com/labstack/echo/v4"
)

func TestHealthHandler(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	if err := HealthHandler(e.NewContext(req, rec)); err != nil {
		t.Fatalf("health handler: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Fatalf("unexpected response body: %s", rec.Body.String())
	}
}

func TestReadinessHandler_UninitializedDependencies(t *testing.T) {
	oldDB := app.DB
	oldRabbit := app.Rabbit
	app.DB = nil
	app.Rabbit = nil
	t.Cleanup(func() {
		app.DB = oldDB
		app.Rabbit = oldRabbit
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	if err := ReadinessHandler(e.NewContext(req, rec)); err != nil {
		t.Fatalf("readiness handler: %v", err)
	}

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"status":"not_ready"`) {
		t.Fatalf("unexpected response body: %s", rec.Body.String())
	}
}
