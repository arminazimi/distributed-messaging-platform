package health

import (
	"context"
	"net/http"
	"time"

	"gateway/app"

	"github.com/labstack/echo/v4"
)

const readinessTimeout = 2 * time.Second

type response struct {
	Status       string            `json:"status"`
	Dependencies map[string]string `json:"dependencies,omitempty"`
}

func HealthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, response{Status: "ok"})
}

func ReadinessHandler(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), readinessTimeout)
	defer cancel()

	dependencies := map[string]string{}

	if err := checkMySQL(ctx); err != nil {
		dependencies["mysql"] = "unavailable"
		return c.JSON(http.StatusServiceUnavailable, response{
			Status:       "not_ready",
			Dependencies: dependencies,
		})
	}
	dependencies["mysql"] = "ok"

	if err := checkRabbitMQ(); err != nil {
		dependencies["rabbitmq"] = "unavailable"
		return c.JSON(http.StatusServiceUnavailable, response{
			Status:       "not_ready",
			Dependencies: dependencies,
		})
	}
	dependencies["rabbitmq"] = "ok"

	return c.JSON(http.StatusOK, response{
		Status:       "ready",
		Dependencies: dependencies,
	})
}

func checkMySQL(ctx context.Context) error {
	if app.DB == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "mysql is not initialized")
	}
	return app.DB.PingContext(ctx)
}

func checkRabbitMQ() error {
	if app.Rabbit == nil || app.Rabbit.Conn == nil || app.Rabbit.Conn.IsClosed() {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "rabbitmq is not initialized")
	}

	ch, err := app.Rabbit.Conn.Channel()
	if err != nil {
		return err
	}
	return ch.Close()
}
