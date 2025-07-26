// Package bridge provides compatibility bridges for third-party libraries
package bridge

import (
	"database/sql"
	"time"

	"github.com/derailed/k9s/internal/migrate"
	"github.com/go-gorp/gorp/v3"
)

// Dialects is a map of dialect strings to dialect implementations
var Dialects = map[string]gorp.Dialect{
	"sqlite3":   gorp.SqliteDialect{},
	"postgres":  gorp.PostgresDialect{},
	"mysql":     gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8"},
	"mssql":     gorp.SqlServerDialect{},
	"oci8":      migrate.FixedOracleDialect{},
	"godror":    migrate.FixedOracleDialect{},
	"snowflake": migrate.FixedSnowflakeDialect{},
}

// SetDialects replaces the dialect map with the provided one
func SetDialects(dialects map[string]gorp.Dialect) {
	Dialects = dialects
}

// This file provides a compatibility layer for helm.sh/helm/v3/pkg/storage/driver
// which depends on github.com/rubenv/sql-migrate.
// We implement enough of the sql-migrate API to satisfy the Helm dependencies.

// ExecMax executes a query with args, ensuring that the changes
// are limited by max.
func ExecMax(db *sql.DB, dialect string, max int, query string, args ...interface{}) (sql.Result, error) {
	return migrate.ExecMax(db, dialect, max, query, args...)
}

// NewDbMap returns a new DbMap using the given database connection and dialect.
func NewDbMap(db *sql.DB, dialect string) (*gorp.DbMap, error) {
	return migrate.NewDbMap(db, dialect)
}

// GetDialect returns the dialect for the given driver string
func GetDialect(driver string) gorp.Dialect {
	return migrate.GetDialect(driver)
}

// A MigrationSource interface is used to provide the actual migration sources.
type MigrationSource interface {
	migrate.MigrationSource
}

// FileMigrationSource implements MigrationSource
// for migrations from a directory.
type FileMigrationSource struct {
	Dir string
}

// FindMigrations returns all migrations in the source directory.
func (f FileMigrationSource) FindMigrations() ([]*migrate.Migration, error) {
	src := migrate.FileMigrationSource{Dir: f.Dir}
	return src.FindMigrations()
}

// MemoryMigrationSource implements MigrationSource
// for migrations held in memory.
type MemoryMigrationSource struct {
	Migrations []*migrate.Migration
}

// FindMigrations returns all migrations from the source.
func (m MemoryMigrationSource) FindMigrations() ([]*migrate.Migration, error) {
	src := migrate.MemoryMigrationSource{Migrations: m.Migrations}
	return src.FindMigrations()
}

// MigrationRecord is a record of a migration from a database.
type MigrationRecord struct {
	ID        string    `db:"id"`
	AppliedAt time.Time `db:"applied_at"`
}

// MigrationSet provides a collection of migrations.
type MigrationSet struct {
	TableName              string
	SchemaName             string
	Migrations             []*migrate.Migration
	Dialect                gorp.Dialect
	Direction              migrate.Direction
	DisableTransactionUp   bool
	DisableTransactionDown bool
}

// Direction is the direction in which migrations should be run.
type Direction migrate.Direction

const (
	// Up migrations should be run
	Up Direction = Direction(migrate.Up)
	// Down migrations should be run
	Down Direction = Direction(migrate.Down)
)

// Migrate runs a migration.
func Migrate(db *sql.DB, dialect string, migrations MigrationSource, dir Direction, max int) (int, error) {
	return migrate.Migrate(db, dialect, migrations, migrate.Direction(dir), max)
}
