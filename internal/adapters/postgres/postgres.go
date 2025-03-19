package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/Olegsandrik/Exponenta/config"
	"github.com/Olegsandrik/Exponenta/logger"

	"github.com/jmoiron/sqlx"

	// Используется для регистрации драйвера PostgreSQL (pgx).
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Adapter struct {
	db *sqlx.DB
}

func NewPostgresAdapter(config *config.Config) (*Adapter, error) {
	psqlInfo := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s&search_path=%s",
		config.PostgresUser,
		config.PostgresPasswd,
		config.PostgresEndpoint,
		config.PostgresPort,
		config.PostgresDBName,
		config.PostgresDisable,
		config.PostgresPublic,
	)

	db, err := sqlx.Connect(config.PostgresDriverName, psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}

	db.SetMaxOpenConns(config.PostgresMaxOpenConn)
	db.SetConnMaxIdleTime(config.PostgresConnIdleTime)

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

func (a *Adapter) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return a.db.QueryContext(ctx, query, args...)
}

func (a *Adapter) QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	return a.db.QueryxContext(ctx, query, args...)
}

func (a *Adapter) QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	return a.db.QueryRowxContext(ctx, query, args...)
}
