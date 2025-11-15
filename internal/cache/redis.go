package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vibe-gaming/backend/internal/config"
)

const (
	RedisTypeSingle  = "redis"
	RedisTypeCluster = "redisCluster"
	pingTimeout      = time.Millisecond * 1500
)

func NewRedis(cfg config.Cache) (redis.UniversalClient, error) {
	if cfg.Type == RedisTypeSingle {
		return newRedis(cfg)
	}
	if cfg.Type == RedisTypeCluster {
		return newRedisCluster(cfg)
	}

	return nil, errors.New("wrong redis type")
}

func newRedis(cfg config.Cache) (*redis.Client, error) {
	opts := &redis.Options{
		Addr:            cfg.Redis.Address,
		Password:        "",
		DB:              0,
		PoolSize:        2000,
		ConnMaxIdleTime: 170 * time.Second,
		DialTimeout:     time.Second * 1,
		ReadTimeout:     time.Second * 1,
		WriteTimeout:    time.Second * 1,
	}
	client := redis.NewClient(opts)

	pingCtx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	_, err := client.Ping(pingCtx).Result()
	return client, err
}

func newRedisCluster(cfg config.Cache) (*redis.ClusterClient, error) {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:           cfg.RedisCluster.Addresses,
		RouteRandomly:   false, // send read operations only to master nodes
		ReadOnly:        false, // send read operations only to master nodes
		PoolSize:        100,
		ConnMaxLifetime: 15 * time.Minute,
		DialTimeout:     time.Second * 1,
		ReadTimeout:     time.Second * 1,
		WriteTimeout:    time.Second * 1,
	})

	pingCtx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	_, err := client.Ping(pingCtx).Result()

	return client, err
}
