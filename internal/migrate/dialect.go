// Package migrate provides database migration utilities
// This is a drop-in replacement for github.com/rubenv/sql-migrate
// that fixes compatibility issues with new versions of gorp
package migrate

import (
	"database/sql"
	"fmt"
	"regexp"
	"time"

	"github.com/go-gorp/gorp/v3"
)

// Direction is the direction in which migrations should be applied
type Direction int

const (
	// Up migrations should be applied
	Up Direction = iota
	// Down migrations should be reversed
	Down
)

// FixedSnowflakeDialect is a dialect for Snowflake databases with SetTableNameMapper
type FixedSnowflakeDialect struct {
	gorp.SnowflakeDialect
}

// SetTableNameMapper implements the required method for gorp.Dialect
func (d FixedSnowflakeDialect) SetTableNameMapper(mapper interface{}) {
	// No-op implementation to satisfy the interface
}

// FixedOracleDialect implements OracleDialect with SetTableNameMapper
type FixedOracleDialect struct {
	gorp.OracleDialect
}

// SetTableNameMapper implements the required method for gorp.Dialect
func (d FixedOracleDialect) SetTableNameMapper(mapper interface{}) {
	// No-op implementation to satisfy the interface
}

// MigrationSet defines a collection of migrations
type MigrationSet struct {
	TableName              string
	SchemaName             string
	Migrations             []*Migration
	Dialect                gorp.Dialect
	Direction              Direction
	DisableTransactionUp   bool
	DisableTransactionDown bool
}

// Migration defines a database migration
type Migration struct {
	Id       string
	Up       []string
	Down     []string
	Checksum []byte
}

// MigrationType is the type of migration
type MigrationType string

// The possible migration types
const (
	MigrationTypeSQL     MigrationType = "sql"
	MigrationTypeGo      MigrationType = "go"
	MigrationTypeUnknown MigrationType = "unknown"
)

// MigrationSource interface defines a source of migration scripts
type MigrationSource interface {
	FindMigrations() ([]*Migration, error)
}

// MemoryMigrationSource implements MigrationSource for migrations defined in memory
type MemoryMigrationSource struct {
	Migrations []*Migration
}

// FindMigrations returns the defined migrations
func (m MemoryMigrationSource) FindMigrations() ([]*Migration, error) {
	return m.Migrations, nil
}

// MigrationRecord represents a previously executed migration from the database
type MigrationRecord struct {
	ID        string    `db:"id"`
	AppliedAt time.Time `db:"applied_at"`
}

// NewDbMap creates a new DbMap object for the given database connection
func NewDbMap(db *sql.DB, dialect string) (*gorp.DbMap, error) {
	var d gorp.Dialect
	switch dialect {
	case "sqlite3":
		d = gorp.SqliteDialect{}
	case "postgres":
		d = gorp.PostgresDialect{}
	case "mysql":
		d = gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8"}
	case "mssql":
		d = gorp.SqlServerDialect{}
	case "oci8", "godror":
		d = FixedOracleDialect{}
	case "snowflake":
		d = FixedSnowflakeDialect{}
	default:
		return nil, fmt.Errorf("unsupported dialect: %s", dialect)
	}

	return &gorp.DbMap{Db: db, Dialect: d}, nil
}

// ExecMax executes the given SQL using the provided database connection
func ExecMax(db *sql.DB, dialect string, max int, query string, args ...interface{}) (sql.Result, error) {
	switch dialect {
	case "postgres", "mysql":
		query = regexp.MustCompile(`\$\d+`).ReplaceAllString(query, "?")
	}
	return db.Exec(query, args...)
}

// GetDialect returns the dialect object for the given dialect string
func GetDialect(dialect string) gorp.Dialect {
	switch dialect {
	case "sqlite3":
		return gorp.SqliteDialect{}
	case "postgres":
		return gorp.PostgresDialect{}
	case "mysql":
		return gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8"}
	case "mssql":
		return gorp.SqlServerDialect{}
	case "oci8", "godror":
		return FixedOracleDialect{}
	case "snowflake":
		return FixedSnowflakeDialect{}
	}
	return nil
}
