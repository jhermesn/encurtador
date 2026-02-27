package config

import (
	"fmt"
	"os"
)

type Config struct {
	MySQLDSN          string
	RedisAddr         string
	RedisPassword     string
	AppPort           string
	BaseURL           string
	CORSAllowedOrigin string
	FrontendURL       string
}

func Load() (*Config, error) {
	cfg := &Config{
		MySQLDSN:          os.Getenv("MYSQL_DSN"),
		RedisAddr:         os.Getenv("REDIS_ADDR"),
		RedisPassword:     os.Getenv("REDIS_PASSWORD"),
		AppPort:           os.Getenv("APP_PORT"),
		BaseURL:           os.Getenv("BASE_URL"),
		CORSAllowedOrigin: os.Getenv("CORS_ALLOWED_ORIGIN"),
		FrontendURL:       os.Getenv("FRONTEND_URL"),
	}

	if cfg.MySQLDSN == "" {
		return nil, fmt.Errorf("MYSQL_DSN is required")
	}
	if cfg.RedisAddr == "" {
		return nil, fmt.Errorf("REDIS_ADDR is required")
	}
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("BASE_URL is required")
	}
	if cfg.AppPort == "" {
		cfg.AppPort = "8080"
	}
	if cfg.CORSAllowedOrigin == "" {
		return nil, fmt.Errorf("CORS_ALLOWED_ORIGIN is required")
	}
	if cfg.FrontendURL == "" {
		return nil, fmt.Errorf("FRONTEND_URL is required")
	}

	return cfg, nil
}
