package database

import (
	"context"
	"database/sql"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joomcode/errorx"
)

var dbConn *sql.DB
var m sync.Mutex

// pingTimeout bounds how long the readiness probe will wait for the database
// to respond before declaring the connection unhealthy. Keep it short — the
// probe is hit on every readyz request and shouldn't stall request handling.
const pingTimeout = 1 * time.Second

// Connection returns the shared *sql.DB singleton, or nil if Connect has not
// been called or DSN was empty.
func Connection() *sql.DB {
	m.Lock()
	defer m.Unlock()
	return dbConn
}

// Connect opens a new database connection using the driver and DSN in cfg,
// verifies reachability with a ping, and stores the result in the package
// singleton. Returns nil (a no-op) when cfg.Enabled() is false or when a
// connection is already established. Returns an error if the connection
// cannot be opened or pinged.
func Connect(cfg *Config) error {
	if !cfg.Enabled() {
		return nil
	}

	m.Lock()
	defer m.Unlock()

	if dbConn != nil {
		return nil
	}

	conn, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return errorx.InitializationFailed.Wrap(err, "failed to open database connection")
	}

	conn.SetMaxOpenConns(cfg.MaxOpenConns)
	conn.SetMaxIdleConns(cfg.MaxIdleConns)
	conn.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	conn.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	if err = conn.Ping(); err != nil {
		_ = conn.Close()
		return errorx.InitializationFailed.Wrap(err, "failed to ping database")
	}

	dbConn = conn
	return nil
}

// Disconnect closes the shared database connection and resets the singleton
// to nil. It is a no-op if no connection is established.
func Disconnect() error {
	m.Lock()
	defer m.Unlock()

	if dbConn == nil {
		return nil
	}

	if err := dbConn.Close(); err != nil {
		return errorx.InternalError.Wrap(err, "failed to close database connection")
	}

	dbConn = nil
	return nil
}

// IsHealthy returns true if a connection has been established AND a short
// PingContext succeeds. Used by the readiness probe to gate /readyz on
// database availability. Returns false (without erroring) when the database
// subsystem is disabled or the singleton hasn't been initialised yet.
func IsHealthy() bool {
	conn := Connection()
	if conn == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	return conn.PingContext(ctx) == nil
}
