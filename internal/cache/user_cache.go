package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	sqlc "github.com/tfenng/scaffold/internal/gen/sqlc"
)

type UserCache struct {
	Rdb *redis.Client
	TTL time.Duration
}

func NewUserCache(rdb *redis.Client) *UserCache {
	return &UserCache{Rdb: rdb, TTL: 5 * time.Minute}
}

func (c *UserCache) key(id int64) string { return fmt.Sprintf("user:v1:id:%d", id) }

func (c *UserCache) Get(ctx context.Context, id int64) (sqlc.User, bool, error) {
	val, err := c.Rdb.Get(ctx, c.key(id)).Result()
	if err == redis.Nil {
		return sqlc.User{}, false, nil
	}
	if err != nil {
		return sqlc.User{}, false, err
	}
	var u sqlc.User
	if err := json.Unmarshal([]byte(val), &u); err != nil {
		return sqlc.User{}, false, err
	}
	return u, true, nil
}

func (c *UserCache) Set(ctx context.Context, u sqlc.User) error {
	b, _ := json.Marshal(u)
	return c.Rdb.Set(ctx, c.key(u.ID), b, c.TTL).Err()
}

func (c *UserCache) Del(ctx context.Context, id int64) error {
	return c.Rdb.Del(ctx, c.key(id)).Err()
}
