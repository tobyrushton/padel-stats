package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] ")

		if _, err := db.Exec(`ALTER TABLE games ADD COLUMN IF NOT EXISTS creator_id BIGINT`); err != nil {
			return err
		}

		if _, err := db.Exec(`UPDATE games SET creator_id = team1_player1_id WHERE creator_id IS NULL`); err != nil {
			return err
		}

		if _, err := db.Exec(`ALTER TABLE games ALTER COLUMN creator_id SET NOT NULL`); err != nil {
			return err
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] ")

		if _, err := db.Exec(`ALTER TABLE games DROP COLUMN IF EXISTS creator_id`); err != nil {
			return err
		}

		return nil
	})
}
