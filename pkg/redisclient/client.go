package redisclient

import (
	"fmt"
	"net"
	"strings"
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
// Sends AUTH if a password is configured, then PING to verify.
func New(cfg config.RedisConfig) (*Client, error) {
	addr := net.JoinHostPort(cfg.Host, cfg.Port)

	conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil {
		return nil, fmt.Errorf("redis connect %s: %w", addr, err)
	}
	defer conn.Close()

	// AUTH if password is set
	if cfg.Password != "" {
		authCmd := fmt.Sprintf("*2\r\n$4\r\nAUTH\r\n$%d\r\n%s\r\n", len(cfg.Password), cfg.Password)
		if _, err := conn.Write([]byte(authCmd)); err != nil {
			return nil, fmt.Errorf("redis auth write: %w", err)
		}

		buf := make([]byte, 64)
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		n, err := conn.Read(buf)
		if err != nil {
			return nil, fmt.Errorf("redis auth read: %w", err)
		}
		resp := string(buf[:n])
		if !strings.HasPrefix(resp, "+OK") {
			return nil, fmt.Errorf("redis auth failed: %s", strings.TrimSpace(resp))
		}
	}

	// PING
	if _, err := conn.Write([]byte("*1\r\n$4\r\nPING\r\n")); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	buf := make([]byte, 64)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := conn.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("redis ping read: %w", err)
	}

	resp := string(buf[:n])
	if !strings.HasPrefix(resp, "+PONG") {
		return nil, fmt.Errorf("redis unexpected response: %s", strings.TrimSpace(resp))
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
