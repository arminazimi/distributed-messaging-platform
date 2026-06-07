package metrics

import (
	"errors"
	"strconv"

	"github.com/labstack/echo/v4"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequests = prom.NewCounterVec(
		prom.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests processed",
		},
		[]string{"path", "method", "status"},
	)
	httpDuration = prom.NewHistogramVec(
		prom.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency",
			Buckets: prom.DefBuckets,
		},
		[]string{"path", "method"},
	)
	rateLimitedRequests = prom.NewCounterVec(
		prom.CounterOpts{
			Name: "rate_limited_requests_total",
			Help: "Total requests rejected by rate limiting",
		},
		[]string{"scope"},
	)
)

func init() {
	prom.MustRegister(httpRequests, httpDuration, rateLimitedRequests)
}

func RecordRateLimited(scope string) {
	rateLimitedRequests.WithLabelValues(scope).Inc()
}

func EchoMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Path()
			method := c.Request().Method
			timer := prom.NewTimer(httpDuration.WithLabelValues(path, method))
			err := next(c)
			timer.ObserveDuration()
			status := c.Response().Status
			if err != nil {
				status = statusFromError(err)
			}
			if status == 0 {
				status = 200
			}
			httpRequests.WithLabelValues(path, method, strconv.Itoa(status)).Inc()
			return err
		}
	}
}

func statusFromError(err error) int {
	var httpErr *echo.HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.Code
	}
	return 500
}

func Handler() echo.HandlerFunc {
	h := promhttp.Handler()
	return func(c echo.Context) error {
		h.ServeHTTP(c.Response(), c.Request())
		return nil
	}
}
