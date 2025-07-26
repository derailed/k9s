package migrate

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/go-gorp/gorp/v3"
)

// Migrate executes a migration using a DB instance
func Migrate(db *sql.DB, dialect string, migrations MigrationSource, dir Direction, max int) (int, error) {
	dbMap, err := NewDbMap(db, dialect)
	if err != nil {
		return 0, err
	}
	
	return ExecMigration(dbMap, migrations, dir, max)
}

// ExecMigration executes a migration using a DbMap instance
func ExecMigration(dbMap *gorp.DbMap, migrations MigrationSource, dir Direction, max int) (int, error) {
	ms := MigrationSet{
		TableName: "migrations",
		Dialect:   dbMap.Dialect,
		Direction: dir,
	}

	// Find all migrations
	allMigrations, err := migrations.FindMigrations()
	if err != nil {
		return 0, err
	}
	ms.Migrations = allMigrations

	// Make sure the migration table exists
	err = createMigrationTable(dbMap, ms.TableName)
	if err != nil {
		return 0, err
	}

	// Find already executed migrations
	executed, err := findExecutedMigrations(dbMap, ms.TableName)
	if err != nil {
		return 0, err
	}

	// Find migrations that need to be applied
	toApply := findMigrationsToApply(&ms, executed, dir, max)
	
	// Apply the migrations
	applied := 0
	for _, migration := range toApply {
		var queries []string
		if dir == Up {
			queries = migration.Up
		} else {
			queries = migration.Down
		}

		// Begin transaction
		var txn *gorp.Transaction
		if !ms.DisableTransactionUp && dir == Up || !ms.DisableTransactionDown && dir == Down {
			txn, err = dbMap.Begin()
			if err != nil {
				return applied, err
			}
		}

		// Execute each query
		for _, query := range queries {
			if strings.TrimSpace(query) == "" {
				continue
			}

			if txn != nil {
				_, err = txn.Exec(query)
			} else {
				_, err = dbMap.Exec(query)
			}

			if err != nil {
				if txn != nil {
					txn.Rollback()
				}
				return applied, fmt.Errorf("error executing migration %s: %s", migration.Id, err)
			}
		}

		// Record the migration
		if dir == Up {
			err = recordMigration(txn, dbMap, ms.TableName, migration.Id)
			if err != nil {
				if txn != nil {
					txn.Rollback()
				}
				return applied, err
			}
		} else {
			err = removeMigration(txn, dbMap, ms.TableName, migration.Id)
			if err != nil {
				if txn != nil {
					txn.Rollback()
				}
				return applied, err
			}
		}

		// Commit transaction
		if txn != nil {
			err = txn.Commit()
			if err != nil {
				return applied, err
			}
		}

		applied++
	}

	return applied, nil
}

// Create the migration table if it doesn't exist
func createMigrationTable(dbMap *gorp.DbMap, tableName string) error {
	_, err := dbMap.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP NOT NULL,
			PRIMARY KEY (id)
		)
	`, tableName))
	return err
}

// Find migrations that have already been executed
func findExecutedMigrations(dbMap *gorp.DbMap, tableName string) (map[string]struct{}, error) {
	rows, err := dbMap.Query(fmt.Sprintf("SELECT id FROM %s", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	executed := make(map[string]struct{})
	for rows.Next() {
		var id string
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		executed[id] = struct{}{}
	}
	return executed, nil
}

// Find migrations that need to be applied
func findMigrationsToApply(set *MigrationSet, executed map[string]struct{}, direction Direction, max int) []*Migration {
	var result []*Migration
	
	// Sort migrations based on ID
	// For simplicity, we'll assume they're already sorted
	
	for _, migration := range set.Migrations {
		if direction == Up {
			// If it's already executed, skip it
			if _, ok := executed[migration.Id]; ok {
				continue
			}
		} else {
			// If it's not executed, skip it
			if _, ok := executed[migration.Id]; !ok {
				continue
			}
		}
		
		// Add to the list
		result = append(result, migration)
		
		// Stop if we've reached the max
		if max > 0 && len(result) >= max {
			break
		}
	}
	
	return result
}

// Record a migration as executed
func recordMigration(txn *gorp.Transaction, dbMap *gorp.DbMap, tableName, id string) error {
	query := fmt.Sprintf("INSERT INTO %s (id, applied_at) VALUES (?, ?)", tableName)
	if dbMap.Dialect.BindVar(1) == "$1" {
		query = strings.Replace(query, "?", "$1", 1)
		query = strings.Replace(query, "?", "$2", 1)
	}
	
	if txn != nil {
		_, err := txn.Exec(query, id, time.Now())
		return err
	}
	_, err := dbMap.Exec(query, id, time.Now())
	return err
}

// Remove a migration from the executed list
func removeMigration(txn *gorp.Transaction, dbMap *gorp.DbMap, tableName, id string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", tableName)
	if dbMap.Dialect.BindVar(1) == "$1" {
		query = strings.Replace(query, "?", "$1", 1)
	}
	
	if txn != nil {
		_, err := txn.Exec(query, id)
		return err
	}
	_, err := dbMap.Exec(query, id)
	return err
}
