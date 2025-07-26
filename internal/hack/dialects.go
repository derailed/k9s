// Package hack provides workarounds for various issues
package hack

import (
	"github.com/go-gorp/gorp/v3"
)

// MigrationDialects is a map of dialect names to dialect implementations
var MigrationDialects map[string]gorp.Dialect

// FixedOracleDialect implements OracleDialect with SetTableNameMapper
type FixedOracleDialect struct {
	gorp.OracleDialect
}

// SetTableNameMapper implements the required method for gorp.Dialect
func (d FixedOracleDialect) SetTableNameMapper(mapper interface{}) {
	// No-op implementation to satisfy the interface
}

// FixedSnowflakeDialect is a dialect for Snowflake databases with SetTableNameMapper
type FixedSnowflakeDialect struct {
	gorp.SnowflakeDialect
}

// SetTableNameMapper implements the required method for gorp.Dialect
func (d FixedSnowflakeDialect) SetTableNameMapper(mapper interface{}) {
	// No-op implementation to satisfy the interface
}

// InitDialects initializes the dialect map
func InitDialects() {
	MigrationDialects = map[string]gorp.Dialect{
		"sqlite3":   gorp.SqliteDialect{},
		"postgres":  gorp.PostgresDialect{},
		"mysql":     gorp.MySQLDialect{Engine: "InnoDB", Encoding: "UTF8"},
		"mssql":     gorp.SqlServerDialect{},
		"oci8":      FixedOracleDialect{},
		"godror":    FixedOracleDialect{},
		"snowflake": FixedSnowflakeDialect{},
	}
}
