package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
)

func NewPostgres(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	config.MaxConns = int32(viper.GetInt("postgres.max_conns"))
	config.MinConns = int32(viper.GetInt("postgres.min_conns"))

	pool, err := pgxpool.NewWithConfig(
		ctx,
		config,
	)

	if err != nil {
		return nil, err
	}

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
