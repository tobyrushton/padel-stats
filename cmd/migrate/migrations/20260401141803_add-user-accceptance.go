package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] ")

		_, err := db.ExecContext(ctx, `
			ALTER TABLE users
			ADD COLUMN is_accepted_by_admin BOOLEAN NOT NULL DEFAULT FALSE;
		`)

		return err
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] ")
		_, err := db.ExecContext(ctx, `
			ALTER TABLE users
			DROP COLUMN is_accepted_by_admin;
		`)

		return err
	})
}
