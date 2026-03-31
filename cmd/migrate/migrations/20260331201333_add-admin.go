package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] ")

		if _, err := db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS is_admin BOOLEAN DEFAULT FALSE`); err != nil {
			return err
		}

		if _, err := db.Exec(`UPDATE users SET is_admin = FALSE WHERE is_admin IS NULL`); err != nil {
			return err
		}

		if _, err := db.Exec(`ALTER TABLE users ALTER COLUMN is_admin SET NOT NULL`); err != nil {
			return err
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] ")

		if _, err := db.Exec(`ALTER TABLE users DROP COLUMN IF EXISTS is_admin`); err != nil {
			return err
		}

		return nil
	})
}
