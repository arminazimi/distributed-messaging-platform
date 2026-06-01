package main

import (
	"context"
	"gateway/app"
	"gateway/config"
	"gateway/internal/balance"
	"gateway/internal/health"
	"gateway/internal/message"
	"gateway/pkg/metrics"
	"os/signal"
	"syscall"
	"time"

	_ "gateway/docs"

	echSwagger "github.com/swaggo/echo-swagger"
)

// @title           Distributed Messaging Platform API
// @version         1.0
// @description     Distributed messaging platform with balance management, asynchronous delivery, and operator failover.
// @host            localhost:8080
// @BasePath        /
func main() {
	app.Init()

	// Handlers
	app.Echo.POST("/messages/send", message.SendHandler)
	app.Echo.GET("/messages/history", message.HistoryHandler)

	app.Echo.GET("/balance", balance.GetBalanceAndHistoryHandler)
	app.Echo.POST("/balance/add", balance.AddBalanceHandler)

	app.Echo.GET("/healthz", health.HealthHandler)
	app.Echo.GET("/readyz", health.ReadinessHandler)

	app.Echo.GET("/swagger/*", echSwagger.WrapHandler)
	app.Echo.GET("/metrics", metrics.Handler())

	// Graceful ShoutDown
	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- app.Echo.Start(config.AppListenAddr)
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	consumerErrCh := make(chan error, 1)
	go func() {
		consumerErrCh <- message.StartConsumers(ctx)
	}()

	outboxErrCh := make(chan error, 1)
	go func() {
		outboxErrCh <- message.StartOutboxPublisher(ctx)
	}()

	select {
	case err := <-consumerErrCh:
		if err != nil {
			app.Logger.Error("consumer error", "err", err)
		}
	case err := <-outboxErrCh:
		if err != nil {
			app.Logger.Error("outbox error", "err", err)
		}
	case err := <-serverErrCh:
		if err != nil {
			app.Logger.Error("server error", "err", err)
		}
	case <-ctx.Done():
		app.Logger.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.Echo.Shutdown(shutdownCtx); err != nil {
		app.Logger.Error("echo shutdown", "err", err)
	}

	stop()
	app.Shutdown()
}
