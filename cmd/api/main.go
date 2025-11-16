package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kreezerit/pr-assigner/internal/app/handler"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/kreezerit/pr-assigner/internal/app/config"
	prHandler "github.com/kreezerit/pr-assigner/internal/app/handler/pullrequest"
	statsHandler "github.com/kreezerit/pr-assigner/internal/app/handler/statistics"
	teamHandler "github.com/kreezerit/pr-assigner/internal/app/handler/team"
	userHandler "github.com/kreezerit/pr-assigner/internal/app/handler/user"
	"github.com/kreezerit/pr-assigner/internal/domain/pullrequest"
	"github.com/kreezerit/pr-assigner/internal/domain/statistics"
	"github.com/kreezerit/pr-assigner/internal/domain/team"
	"github.com/kreezerit/pr-assigner/internal/domain/user"
	"github.com/kreezerit/pr-assigner/internal/infrastructure/logger"
	"github.com/kreezerit/pr-assigner/internal/infrastructure/persistence/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	if err := logger.Initialize(cfg.Logger.Level); err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	logger.Info("starting PR reviewer service", zap.String("port", cfg.Server.Port))

	db, err := connectDB(cfg.Database)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("failed to close database connection", zap.Error(err))
		}
	}()

	logger.Info("success database connected")

	teamRepo := postgres.NewTeamRepository(db)
	userRepo := postgres.NewUserRepository(db)
	prRepo := postgres.NewPullRequestRepository(db)
	statsRepo := postgres.NewStatisticsRepository(db)

	teamService := team.NewService(teamRepo, userRepo)
	userService := user.NewService(userRepo)
	prService := pullrequest.NewService(prRepo, userRepo)
	statsService := statistics.NewService(statsRepo)

	e := buildServer(teamService, userService, prService, statsService)

	// запускаем HTTP‑сервер в отдельной горутине, чтобы основная горутина могла ждать сигналов завершения
	go func() {
		if err := e.Start(":" + cfg.Server.Port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("failed to start server", zap.Error(err))
		}
	}()

	logger.Info("server started", zap.String("port", cfg.Server.Port))

	// ожидаем системный сигнал (Ctrl+C, docker stop)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server")

	// плавная остановка сервера, до 10 секунд на обработку активных запросов
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", zap.Error(err))
	}

	logger.Info("server stopped")
}

func buildServer(
	teamService team.Service,
	userService user.Service,
	prService pullrequest.Service,
	statsService statistics.Service,
) *echo.Echo {
	teamH := teamHandler.NewHandler(teamService)
	userH := userHandler.NewHandler(userService)
	prH := prHandler.NewHandler(prService)
	statsH := statsHandler.NewHandler(statsService)

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Recover())
	e.Use(handler.LoggingMiddleware)
	e.Use(handler.MetricsMiddleware)
	e.Use(middleware.CORS())

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	teamH.RegisterRoutes(e)
	userH.RegisterRoutes(e)
	prH.RegisterRoutes(e)
	statsH.RegisterRoutes(e)

	return e
}

func connectDB(cfg config.DatabaseConfig) (*sql.DB, error) {
	dsn := cfg.DSN()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	for attempt := 1; attempt <= cfg.MaxRetries; attempt++ {
		if err = db.Ping(); err == nil {
			break
		}

		logger.Warn("failed to ping database",
			zap.Int("attempt", attempt),
			zap.Error(err),
		)

		time.Sleep(cfg.RetryDelay)
	}

	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to connect after %d attempts: %w", cfg.MaxRetries, err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return db, nil
}
