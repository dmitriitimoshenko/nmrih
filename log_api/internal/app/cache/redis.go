package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisClient(addr, password string, db int, defaultTTL time.Duration) *Redis {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &Redis{client: rdb, ttl: defaultTTL}
}

func (r *Redis) Get(ctx context.Context, key string) (string, bool, error) {
	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", false, nil
	}
	return val, err == nil, err
}

func (r *Redis) Set(ctx context.Context, key, value string, ttlOverride *time.Duration) error {
	if ttlOverride != nil {
		return r.client.Set(ctx, key, value, *ttlOverride).Err()
	}
	return r.client.Set(ctx, key, value, r.ttl).Err()
}

func (r *Redis) FlushAll(ctx context.Context) error {
	return r.client.FlushAll(ctx).Err()
}
