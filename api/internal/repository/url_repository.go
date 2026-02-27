package repository

import (
	"context"
	"time"

	"encurtador/internal/model"
)

type URLRepository interface {
	Create(ctx context.Context, url *model.URL) error
	FindBySlug(ctx context.Context, slug string) (*model.URL, error)
	SlugExists(ctx context.Context, slug string) (bool, error)
	ExpireBySlug(ctx context.Context, slug, manageTokenHash string) (bool, error)
	DeleteExpired(ctx context.Context) error
}

type URLCache interface {
	Get(ctx context.Context, slug string) (*model.CachedURL, error)
	Set(ctx context.Context, slug string, cached *model.CachedURL, ttl time.Duration) error
	Delete(ctx context.Context, slug string) error
}
