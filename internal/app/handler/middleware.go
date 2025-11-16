package handler

import (
	"net/http"
	"time"

	"github.com/kreezerit/pr-assigner/internal/infrastructure/logger"
	"github.com/kreezerit/pr-assigner/internal/infrastructure/metrics"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// LoggingMiddleware логирует каждый HTTP‑запрос/ответ (метод, путь, статус, длительность, IP)
func LoggingMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()

		err := next(c)

		duration := time.Since(start)
		path := c.Path()

		logger.Info("http request",
			zap.String("method", c.Request().Method),
			zap.String("path", path),
			zap.Int("status", c.Response().Status),
			zap.Duration("duration", duration),
			zap.String("remote_ip", c.RealIP()),
		)

		return err
	}
}

// MetricsMiddleware записывает Prometheus метрики для HTTP‑запросов (количество, длительность)
func MetricsMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()

		err := next(c)

		duration := time.Since(start).Seconds()
		status := c.Response().Status
		method := c.Request().Method
		path := c.Path()

		metrics.HTTPRequestsTotal.WithLabelValues(
			method,
			path,
			http.StatusText(status),
		).Inc()

		metrics.HTTPRequestDuration.WithLabelValues(
			method,
			path,
		).Observe(duration)

		return err
	}
}
