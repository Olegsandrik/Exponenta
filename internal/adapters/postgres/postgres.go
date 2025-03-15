package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Olegsandrik/Exponenta/config"
	"github.com/Olegsandrik/Exponenta/logger"
	"log/slog"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type Adapter struct {
	db *sqlx.DB
}

func NewPostgresAdapter() (*Adapter, error) {
	psqlInfo := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s&search_path=%s",
		config.POSTGRES_USER,
		config.POSTGRES_PASSWD,
		config.POSTGRES_ENDPOINT,
		config.POSTGRES_PORT,
		config.POSTGRES_DB_NAME,
		config.POSTGRES_DISABLE,
		config.POSTGRES_PUBLIC,
	)

	db, err := sqlx.Connect(config.POSTGRES_DRIVER_NAME, psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	db.SetMaxOpenConns(config.POSTGRES_MAX_OPEN_CONN)
	db.SetConnMaxIdleTime(config.POSTGRES_CONN_IDLE_TIME)

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &Adapter{
		db: db,
	}, nil
}

func (a *Adapter) Close() error {
	return a.db.Close()
}

func (a *Adapter) Exec(ctx context.Context, q string, args ...interface{}) (sql.Result, error) {
	logger.Info(ctx, q, slog.String("args", fmt.Sprintf("%v", args)))
	start := time.Now()
	sqlRes, err := a.db.ExecContext(ctx, sqlx.Rebind(sqlx.DOLLAR, q), args...)
	logger.Info(ctx, fmt.Sprintf("queried in %s", time.Since(start)))
	return sqlRes, err
}

func (a *Adapter) Select(ctx context.Context, dest interface{}, q string, args ...interface{}) error {
	logger.Info(ctx, q, slog.String("args", fmt.Sprintf("%v", args)))
	start := time.Now()
	err := a.db.SelectContext(ctx, dest, sqlx.Rebind(sqlx.DOLLAR, q), args...)
	logger.Info(ctx, fmt.Sprintf("queried in %s", time.Since(start)))
	return err
}

func (a *Adapter) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	logger.Info(ctx, query, slog.String("args", fmt.Sprintf("%v", args)))
	start := time.Now()
	row := a.db.QueryRowContext(ctx, query, args...)
	logger.Info(ctx, fmt.Sprintf("queried in %s", time.Since(start)))
	return row
}

func (a *Adapter) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error) {
	return a.db.BeginTxx(ctx, opts)
}
