package migrate

import (
	"github.com/go-gorp/gorp/v3"
)

// TableNameMapper is the same as gorp.TableNameMapper
type TableNameMapper func(string) string

// SnowflakeDialect wraps gorp.SnowflakeDialect to implement the SetTableNameMapper method
type SnowflakeDialect struct {
	gorp.SnowflakeDialect
}

// SetTableNameMapper implements the method required by gorp.Dialect interface
func (SnowflakeDialect) SetTableNameMapper(mapper any) {
	// Implementation is not needed for sql-migrate usage
}

// OracleDialect wraps gorp.OracleDialect to implement the SetTableNameMapper method
type OracleDialect struct {
	gorp.OracleDialect
}

// SetTableNameMapper implements the method required by gorp.Dialect interface
func (OracleDialect) SetTableNameMapper(mapper any) {
	// Implementation is not needed for sql-migrate usage
}
