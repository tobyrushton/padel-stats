package migrations

import (
	"context"
	"fmt"

	"github.com/tobyrushton/padel-stats/libs/db/models"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] ")

		_, err := db.NewCreateTable().
			Model((*models.Session)(nil)).
			IfNotExists().
			Exec(ctx)
		if err != nil {
			return err
		}

		return nil
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] ")

		_, err := db.NewDropTable().
			Model((*models.Session)(nil)).
			IfExists().
			Exec(ctx)
		if err != nil {
			return err
		}

		return nil
	})
}
