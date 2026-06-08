package backpressure

import (
	"context"
	"net/http"

	"gateway/pkg/metrics"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type Result struct {
	Active       bool
	PendingCount int64
}

type Getter interface {
	GetContext(ctx context.Context, dest any, query string, args ...any) error
}

func Check(ctx context.Context, db Getter, enabled bool, threshold int) (Result, error) {
	if !enabled {
		return Result{}, nil
	}

	pending, err := PendingOutboxEvents(ctx, db)
	if err != nil {
		return Result{}, err
	}
	metrics.SetOutboxPendingEvents(pending)

	return Result{
		Active:       pending > int64(threshold),
		PendingCount: pending,
	}, nil
}

func PendingOutboxEvents(ctx context.Context, db Getter) (int64, error) {
	var count int64
	const query = `
		SELECT COUNT(*)
		FROM outbox_events
		WHERE status = 'pending'
		  AND event_type = 'message.send'
	`
	if err := db.GetContext(ctx, &count, query); err != nil {
		return 0, err
	}
	return count, nil
}

func Respond(c echo.Context) error {
	metrics.RecordBackpressureRejection()
	return c.JSON(http.StatusServiceUnavailable, ErrorResponse{
		Error:   "system_overloaded",
		Message: "The system is temporarily overloaded. Please retry later.",
	})
}

var _ Getter = (*sqlx.DB)(nil)
