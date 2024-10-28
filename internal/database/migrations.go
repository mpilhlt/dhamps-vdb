package database

// This is mostly from https://github.com/tardisx/embed_tern,
// with some adjustments for pgxpool and for how it is called in my app's context.

import (
	"context"
	"embed"
	"fmt"
	"io/fs"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/tern/v2/migrate"
)

const versionTable = "db_version"

type Migrator struct {
  migrator *migrate.Migrator
}

//go:embed migrations/*.sql
var migrationFiles embed.FS

func NewMigrator(ctx context.Context, conn *pgx.Conn) (Migrator, error) {

  migrator, err := migrate.NewMigratorEx(
    ctx, conn, versionTable,
    &migrate.MigratorOptions{
      DisableTx: false,
    })
  if err != nil {
    return Migrator{}, err
  }

  migrationRoot, err := fs.Sub(migrationFiles, "migrations")
  if err != nil {
    return Migrator{}, err
  }

  err = migrator.LoadMigrations(migrationRoot)
  if err != nil {
    return Migrator{}, err
  }

  return Migrator{
    migrator: migrator,
  }, nil
}

// Info the current migration version and the embedded maximum migration, and a textual
// representation of the migration state for informational purposes.
func (m Migrator) Info() (int32, int32, string, error) {

  version, err := m.migrator.GetCurrentVersion(context.Background())
  if err != nil {
    return 0, 0, "", err
  }
  info := ""

  var last int32
  for _, thisMigration := range m.migrator.Migrations {
    last = thisMigration.Sequence

    cur := version == thisMigration.Sequence
    indicator := "  "
    if cur {
      indicator = "->"
    }
    info = info + fmt.Sprintf(
      "%2s %3d %s\n",
      indicator,
      thisMigration.Sequence, thisMigration.Name)
  }

  return version, last, info, nil
}

// Migrate migrates the DB to the most recent version of the schema.
func (m Migrator) Migrate(ctx context.Context) error {
  err := m.migrator.Migrate(ctx)
  return err
}

// MigrateTo migrates to a specific version of the schema. Use '0' to undo all migrations.
func (m Migrator) MigrateTo(ctx context.Context, ver int32) error {
  err := m.migrator.MigrateTo(ctx, ver)
  return err
}
