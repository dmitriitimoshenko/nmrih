package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dmitriitimoshenko/nmrih/log_api/internal/app/cache/config"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisClient(config *config.RedisConfig) *Redis {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})
	return &Redis{client: rdb, ttl: config.DefaultTTL}
}

func (r *Redis) Get(ctx context.Context, key string) (*string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	return &val, nil
}

func (r *Redis) GetWithTimeout(
	ctx context.Context,
	key string,
	cacheTimeout time.Duration,
) (*string, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, cacheTimeout)
	defer cancel()

	cached, err := r.Get(timeoutCtx, key)
	if err != nil {
		fmt.Println("Error getting from cache:", err)
		return nil, err
	}
	return cached, nil
}

func (r *Redis) Set(ctx context.Context, key, value string, ttlOverride *time.Duration) error {
	if ttlOverride != nil {
		return r.client.Set(ctx, key, value, *ttlOverride).Err()
	}
	return r.client.Set(ctx, key, value, r.ttl).Err()
}

func (r *Redis) SetWithTimeout(
	ctx context.Context,
	key, value string,
	ttlOverride *time.Duration,
	cacheTimeout time.Duration,
) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, cacheTimeout)
	defer cancel()

	if ttlOverride != nil {
		return r.client.Set(timeoutCtx, key, value, *ttlOverride).Err()
	}
	return r.client.Set(timeoutCtx, key, value, r.ttl).Err()
}

func (r *Redis) FlushAll(ctx context.Context) error {
	return r.client.FlushAll(ctx).Err()
}
