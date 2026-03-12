package redisclient

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/P0l1-0825/Go-destino/internal/config"
)

// Client wraps go-redis/v9 for use by security components
// and other Redis-backed features.
type Client struct {
	rdb  *redis.Client
	addr string
}

// New creates a Redis client and verifies connectivity via PING.
func New(cfg config.RedisConfig) (*Client, error) {
	addr := net.JoinHostPort(cfg.Host, cfg.Port)

	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		rdb.Close()
		return nil, fmt.Errorf("redis connect %s: %w", addr, err)
	}

	return &Client{rdb: rdb, addr: addr}, nil
}

// Addr returns the Redis server address.
func (c *Client) Addr() string {
	return c.addr
}

// Ping checks if Redis is reachable.
func (c *Client) Ping() error {
	return c.rdb.Ping(context.Background()).Err()
}

// Close releases the underlying connection pool.
func (c *Client) Close() error {
	return c.rdb.Close()
}

// Unwrap returns the underlying go-redis client for use by
// Redis-backed security components and other features.
func (c *Client) Unwrap() *redis.Client {
	return c.rdb
}
