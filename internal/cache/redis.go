package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedis(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         addr,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  600 * time.Millisecond,
		WriteTimeout: 600 * time.Millisecond,
	})
}

func Ping(ctx context.Context, rdb *redis.Client) error { return rdb.Ping(ctx).Err() }
