package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] ")

		if _, err := db.Exec(`ALTER TABLE seasons DROP COLUMN IF EXISTS year`); err != nil {
			return err
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] ")

		if _, err := db.Exec(`ALTER TABLE seasons ADD COLUMN IF NOT EXISTS year INTEGER NOT NULL DEFAULT 2023`); err != nil {
			return err
		}

		return nil
	})
}
