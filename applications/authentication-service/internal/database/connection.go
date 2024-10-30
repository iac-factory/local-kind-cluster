package database

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"authentication-service/internal/library/levels"
)

var Pool atomic.Pointer[pgxpool.Pool]

// dsn represents the postgresql connection string.
//   - https://www.postgresql.org/docs/current/libpq-envars.html
//   - https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-PARAMKEYWORDS
func dsn(ctx context.Context) (v string) {
	host := os.Getenv("PGHOST")

	uri := url.URL{
		Scheme: "postgresql",
		Host:   host,
	}

	if uri.Host == "" {
		const value = "localhost"
		slog.Log(ctx, levels.Info, "Host Environment Variable Not Found - Using Default", slog.String("environment-variable", "PGHOST"), slog.String("value", value))
		uri.Host = value
	}

	timeout := os.Getenv("PGCONNECT_TIMEOUT")
	if timeout == "" {
		const value = "10"
		slog.Log(ctx, levels.Info, "Timeout Environment Variable Not Found - Using Default", slog.String("environment-variable", "PGCONNECT_TIMEOUT"), slog.String("value", value))
		timeout = value
	}

	application := os.Getenv("PGAPPNAME")

	sslmode := os.Getenv("PGSSLMODE")
	root := os.Getenv("PGSSLROOTCERT")

	maxconnections := os.Getenv("PGPOOLMAXCONNECTIONS")
	if maxconnections == "" {
		value, cpu := 4, runtime.NumCPU()
		if value < cpu {
			value = cpu
		}

		maxconnections = strconv.Itoa(value)
	}

	minconnections := os.Getenv("PGPOOLMINCONNECTIONS")
	if minconnections == "" {
		minconnections = strconv.Itoa(1)
	}

	tz := os.Getenv("PGTZ")
	if tz == "" {
		tz = "UTC"
	}

	db := os.Getenv("PGDATABASE")
	if db == "" {
		db = "authentication-service"
	}

	query := uri.Query()

	username := os.Getenv("PGUSER")
	password := os.Getenv("PGPASSWORD")
	port := os.Getenv("PGPORT")

	query.Add("user", username)
	query.Add("password", password)
	query.Add("port", port)
	query.Add("connect_timeout", timeout)

	query.Add("application_name", application)

	query.Add("pool_max_conns", maxconnections)
	query.Add("pool_min_conns", minconnections)

	query.Add("sslmode", sslmode)
	query.Add("sslrootcert", root)

	query.Add("dbname", db)

	for key, values := range query {
		if len(values) >= 1 && strings.TrimSpace(values[0]) == "" {
			query.Del(key)
		}
	}

	uri.RawQuery = query.Encode()

	slog.InfoContext(ctx, "PostgreSQL Connection Metadata", slog.String("database", db), slog.String("username", username), slog.String("application", application), slog.String("port", port), slog.String("hostname", host))

	return uri.String()
}

// Connection establishes a connection to the database using [pgxpool].
//   - If a connection pool does not exist, a new one is created and stored in the pool variable.
func Connection(ctx context.Context) (*pgxpool.Conn, error) {
	if Pool.Load() == nil {
		configuration, e := pgxpool.ParseConfig(dsn(ctx))
		if e != nil {
			slog.ErrorContext(ctx, "Unable to Generate Configuration from DSN String", slog.String("error", e.Error()))
			return nil, e
		}

		instance, e := pgxpool.NewWithConfig(ctx, configuration)
		if e != nil {
			slog.ErrorContext(ctx, "Unable to Establish Pool Connection to Database", slog.String("error", e.Error()))
			return nil, e
		}

		Pool.Store(instance)
	}

	return Pool.Load().Acquire(ctx)
}

// Disconnect closes the transaction and releases the connection back to the pool.
// If `tx` is not nil, it rolls back the transaction and logs any error.
// If `connection` is not nil, it releases the connection back to the pool.
func Disconnect(ctx context.Context, connection *pgxpool.Conn, tx pgx.Tx) {
	if tx != nil {
		e := tx.Rollback(ctx)
		if e != nil && !(errors.Is(e, pgx.ErrTxClosed)) {
			slog.ErrorContext(ctx, "Error Rolling Back Transaction", slog.String("error", e.Error()))
		} else if e != nil && (errors.Is(e, pgx.ErrTxClosed)) {
			slog.DebugContext(ctx, "Successfully Committed Database Transaction")
		} else if e == nil {
			slog.InfoContext(ctx, "Successfully Rolled Back Database Transaction")
		}
	}

	if connection != nil {
		connection.Release()
	}
}
