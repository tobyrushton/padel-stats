package postgres

import (
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/extra/bundebug"
)

func NewDb(connUrl string) (*bun.DB, error) {
	config, err := pgx.ParseConfig(connUrl)
	if err != nil {
		return nil, err
	}
	sqldb := stdlib.OpenDB(*config)

	db := bun.NewDB(sqldb, pgdialect.New())

	db.WithQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))

	return db, nil
}
