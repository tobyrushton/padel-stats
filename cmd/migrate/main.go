package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/tobyrushton/bexbox-pl/cmd/migrate/migrations"
	"github.com/tobyrushton/bexbox-pl/libs/config"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"github.com/uptrace/bun/migrate"
)

type CLI struct {
	Init        InitCmd        `cmd:"" help:"Create migration tables"`
	Migrate     MigrateCmd     `cmd:"" help:"Migrate database"`
	Rollback    RollbackCmd    `cmd:"" help:"Rollback the last migration group"`
	Lock        LockCmd        `cmd:"" help:"Lock migrations"`
	Unlock      UnlockCmd      `cmd:"" help:"Unlock migrations"`
	CreateGo    CreateGoCmd    `cmd:"" name:"create-go" help:"Create Go migration"`
	CreateSQL   CreateSQLCmd   `cmd:"" name:"create-sql" help:"Create up and down SQL migrations"`
	CreateTxSQL CreateTxSQLCmd `cmd:"" name:"create-tx-sql" help:"Create up and down transactional SQL migrations"`
	Status      StatusCmd      `cmd:"" help:"Print migrations status"`
	MarkApplied MarkAppliedCmd `cmd:"" name:"mark-applied" help:"Mark migrations as applied without actually running them"`
}

type InitCmd struct{}
type MigrateCmd struct{}
type RollbackCmd struct{}
type LockCmd struct{}
type UnlockCmd struct{}
type StatusCmd struct{}
type MarkAppliedCmd struct{}

type CreateGoCmd struct {
	Name []string `arg:"" optional:"" help:"Migration name parts"`
}

type CreateSQLCmd struct {
	Name []string `arg:"" optional:"" help:"Migration name parts"`
}

type CreateTxSQLCmd struct {
	Name []string `arg:"" optional:"" help:"Migration name parts"`
}

func main() {
	migrator := newMigrator()
	cli := CLI{}
	ctx := kong.Parse(
		&cli,
		kong.BindTo(context.Background(), (*context.Context)(nil)),
		kong.Bind(migrator),
	)

	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}

func newMigrator() *migrate.Migrator {
	appConfig, err := config.MustLoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(appConfig.DBConnectionString)))
	err = sqldb.Ping()
	if err != nil {
		log.Fatal(err)
	}

	db := bun.NewDB(sqldb, pgdialect.New())
	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithEnabled(false),
		bundebug.FromEnv(),
	))

	templateData := map[string]string{
		"Prefix": "example_",
	}

	return migrate.NewMigrator(db, migrations.Migrations, migrate.WithTemplateData(templateData))
}

func (c *InitCmd) Run(ctx context.Context, migrator *migrate.Migrator) error {
	return migrator.Init(ctx)
}

func (c *MigrateCmd) Run(ctx context.Context, migrator *migrate.Migrator) error {
	if err := migrator.Lock(ctx); err != nil {
		return err
	}
	defer migrator.Unlock(ctx) //nolint:errcheck

	group, err := migrator.Migrate(ctx)
	if err != nil {
		return err
	}
	if group.IsZero() {
		fmt.Printf("there are no new migrations to run (database is up to date)\n")
		return nil
	}
	fmt.Printf("migrated to %s\n", group)
	return nil
}

func (c *RollbackCmd) Run(ctx context.Context, migrator *migrate.Migrator) error {
	if err := migrator.Lock(ctx); err != nil {
		return err
	}
	defer migrator.Unlock(ctx) //nolint:errcheck

	group, err := migrator.Rollback(ctx)
	if err != nil {
		return err
	}
	if group.IsZero() {
		fmt.Printf("there are no groups to roll back\n")
		return nil
	}
	fmt.Printf("rolled back %s\n", group)
	return nil
}

func (c *LockCmd) Run(ctx context.Context, migrator *migrate.Migrator) error {
	return migrator.Lock(ctx)
}

func (c *UnlockCmd) Run(ctx context.Context, migrator *migrate.Migrator) error {
	return migrator.Unlock(ctx)
}

func (c *CreateGoCmd) Run(ctx context.Context, migrator *migrate.Migrator) error {
	name := strings.Join(c.Name, "_")
	mf, err := migrator.CreateGoMigration(ctx, name)
	if err != nil {
		return err
	}
	fmt.Printf("created migration %s (%s)\n", mf.Name, mf.Path)
	return nil
}

func (c *CreateSQLCmd) Run(ctx context.Context, migrator *migrate.Migrator) error {
	name := strings.Join(c.Name, "_")
	files, err := migrator.CreateSQLMigrations(ctx, name)
	if err != nil {
		return err
	}

	for _, mf := range files {
		fmt.Printf("created migration %s (%s)\n", mf.Name, mf.Path)
	}

	return nil
}

func (c *CreateTxSQLCmd) Run(ctx context.Context, migrator *migrate.Migrator) error {
	name := strings.Join(c.Name, "_")
	files, err := migrator.CreateTxSQLMigrations(ctx, name)
	if err != nil {
		return err
	}

	for _, mf := range files {
		fmt.Printf("created transaction migration %s (%s)\n", mf.Name, mf.Path)
	}

	return nil
}

func (c *StatusCmd) Run(ctx context.Context, migrator *migrate.Migrator) error {
	ms, err := migrator.MigrationsWithStatus(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("migrations: %s\n", ms)
	fmt.Printf("unapplied migrations: %s\n", ms.Unapplied())
	fmt.Printf("last migration group: %s\n", ms.LastGroup())
	return nil
}

func (c *MarkAppliedCmd) Run(ctx context.Context, migrator *migrate.Migrator) error {
	group, err := migrator.Migrate(ctx, migrate.WithNopMigration())
	if err != nil {
		return err
	}
	if group.IsZero() {
		fmt.Printf("there are no new migrations to mark as applied\n")
		return nil
	}
	fmt.Printf("marked as applied %s\n", group)
	return nil
}
