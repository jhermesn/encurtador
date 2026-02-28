package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	"encurtador/internal/config"
	"encurtador/internal/handler"
	"encurtador/internal/middleware"
	"encurtador/internal/repository"
	"encurtador/internal/service"
	"encurtador/migrations"
)

const (
	serviceName = "encurtador"

	defaultTrustedProxy = "127.0.0.1"
	apiV1BasePath       = "/api/v1"

	redisPingTimeout = 5 * time.Second
	shutdownTimeout  = 10 * time.Second

	mysqlMaxOpenConns    = 25
	mysqlMaxIdleConns    = 10
	mysqlConnMaxLifetime = 5 * time.Minute
)

func main() {
	slog.SetDefault(newJSONLogger())

	cfg, err := config.Load()
	if err != nil {
		slog.Error("loading config", "error", err)
		os.Exit(1)
	}

	db, err := connectMySQL(cfg.MySQLDSN)
	if err != nil {
		slog.Error("connecting to mysql", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	redisClient, err := connectRedis(cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		slog.Error("connecting to redis", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()

	if err := runMigrations(db); err != nil {
		slog.Error("running migrations", "error", err)
		os.Exit(1)
	}

	repo := repository.NewMySQLURLRepository(db)
	cache := repository.NewRedisURLCache(redisClient)
	svc := service.NewURLService(repo, cache, cfg.BaseURL)
	h := handler.NewURLHandler(svc, cfg.FrontendURL)

	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go svc.RunCleanup(appCtx)

	r := buildRouter(h, cfg.CORSAllowedOrigin, cfg.FrontendURL)

	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: r,
	}

	slog.Info("server starting", "port", cfg.AppPort, "base_url", cfg.BaseURL)
	serverErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		slog.Info("shutting down")
	case err := <-serverErr:
		slog.Error("server error", "error", err)
	}
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
	}
}

func newJSONLogger() *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Key = "timestamp"
			}
			return a
		},
	})
	return slog.New(handler).With("service", serviceName)
}

func buildRouter(h *handler.URLHandler, corsOrigin, frontendURL string) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.SetTrustedProxies([]string{defaultTrustedProxy})

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{corsOrigin},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// A single rate limiter instance is shared across the redirect and unlock
	// routes so that enumeration attempts and password guesses count toward
	// the same per-IP budget.
	rl := middleware.NewRateLimiter()

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, frontendURL)
	})

	api := r.Group(apiV1BasePath)
	{
		api.POST("/urls", h.CreateURL)
		api.GET("/urls/check/:slug", h.CheckSlug)
		api.POST("/urls/:slug/unlock", rl, h.UnlockURL)
		api.POST("/urls/:slug/expire", h.ExpireURL)
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
	}

	r.GET("/:slug", rl, h.RedirectOrGate)

	return r
}

func connectMySQL(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("connecting to mysql: %w", err)
	}
	db.SetMaxOpenConns(mysqlMaxOpenConns)
	db.SetMaxIdleConns(mysqlMaxIdleConns)
	db.SetConnMaxLifetime(mysqlConnMaxLifetime)
	return db, nil
}

func connectRedis(addr, password string) (*redis.Client, error) {
	opts := &redis.Options{Addr: addr}
	if password != "" {
		opts.Password = password
	}
	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), redisPingTimeout)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("pinging redis: %w", err)
	}
	return client, nil
}

// runMigrations executes an idempotent schema bootstrap SQL file.
func runMigrations(db *sqlx.DB) error {
	_, err := db.Exec(migrations.BootstrapSQL)
	if err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}
	return nil
}
