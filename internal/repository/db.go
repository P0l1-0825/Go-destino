package repository

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"github.com/P0l1-0825/Go-destino/internal/config"
)

func NewPostgresDB(cfg config.DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Connection pool tuned for Railway PostgreSQL.
	// Railway provides a single shared Postgres instance; the pool must be
	// conservative enough to avoid exhausting the server's max_connections
	// (typically 100 on the hobby plan) while still being able to absorb
	// burst traffic from concurrent API requests.
	//
	// Formula used:
	//   MaxOpenConns  = 20  (leaves headroom for migrations, admin tools)
	//   MaxIdleConns  = 10  (50 % of MaxOpen; avoids thrashing on medium load)
	//   ConnMaxLifetime = 5 min  (recycles connections before Railway's 15-min idle timeout)
	//   ConnMaxIdleTime = 2 min  (aggressively returns idle conns to reduce server load)
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return db, nil
}
