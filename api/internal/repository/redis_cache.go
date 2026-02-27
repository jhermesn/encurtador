package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"encurtador/internal/model"
)

type redisURLCache struct {
	client *redis.Client
}

func NewRedisURLCache(client *redis.Client) URLCache {
	return &redisURLCache{client: client}
}

func cacheKey(slug string) string {
	return "url:" + slug
}

func (c *redisURLCache) Get(ctx context.Context, slug string) (*model.CachedURL, error) {
	val, err := c.client.Get(ctx, cacheKey(slug)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting from redis: %w", err)
	}

	var cached model.CachedURL
	if err := json.Unmarshal(val, &cached); err != nil {
		return nil, fmt.Errorf("unmarshaling cached url: %w", err)
	}
	return &cached, nil
}

func (c *redisURLCache) Set(ctx context.Context, slug string, cached *model.CachedURL, ttl time.Duration) error {
	data, err := json.Marshal(cached)
	if err != nil {
		return fmt.Errorf("marshaling cached url: %w", err)
	}
	if err := c.client.Set(ctx, cacheKey(slug), data, ttl).Err(); err != nil {
		return fmt.Errorf("setting in redis: %w", err)
	}
	return nil
}

func (c *redisURLCache) Delete(ctx context.Context, slug string) error {
	if err := c.client.Del(ctx, cacheKey(slug)).Err(); err != nil {
		return fmt.Errorf("deleting from redis: %w", err)
	}
	return nil
}
