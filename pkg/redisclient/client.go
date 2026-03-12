package redisclient

import (
	"fmt"
	"net"
	"time"

	"github.com/P0l1-0825/Go-destino/internal/config"
)

// Client is a lightweight Redis connectivity wrapper.
// Uses standard library net for health checks. The in-memory security
// components handle token blacklisting, login limiting, and password resets.
// Upgrade to go-redis/v9 when Redis-backed implementations are needed.
type Client struct {
	addr string
}

// New creates a Redis client and verifies connectivity.
func New(cfg config.RedisConfig) (*Client, error) {
	addr := net.JoinHostPort(cfg.Host, cfg.Port)

	conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil {
		return nil, fmt.Errorf("redis connect %s: %w", addr, err)
	}

	// Send PING
	_, err = conn.Write([]byte("*1\r\n$4\r\nPING\r\n"))
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	buf := make([]byte, 64)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := conn.Read(buf)
	conn.Close()

	if err != nil {
		return nil, fmt.Errorf("redis ping read: %w", err)
	}

	resp := string(buf[:n])
	if resp != "+PONG\r\n" {
		return nil, fmt.Errorf("redis unexpected response: %s", resp)
	}

	return &Client{addr: addr}, nil
}

// Addr returns the Redis server address.
func (c *Client) Addr() string {
	return c.addr
}

// Ping checks if Redis is reachable.
func (c *Client) Ping() error {
	conn, err := net.DialTimeout("tcp", c.addr, 2*time.Second)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}
