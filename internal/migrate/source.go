package migrate

import (
	"io/ioutil"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var migrationPattern = regexp.MustCompile(`^([0-9]+)_(.*)\.sql$`)

// FileMigrationSource implements MigrationSource for migrations stored in files
type FileMigrationSource struct {
	Dir string
}

// FindMigrations searches for migrations in the specified directory
func (f FileMigrationSource) FindMigrations() ([]*Migration, error) {
	migrations := make([]*Migration, 0)

	// Read directory
	files, err := ioutil.ReadDir(f.Dir)
	if err != nil {
		return nil, err
	}

	// Create set for identifying migration files
	migMap := make(map[string]*Migration)

	// Find migration files
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Parse filename
		matches := migrationPattern.FindStringSubmatch(file.Name())
		if matches == nil {
			continue
		}

		id := matches[1]
		direction := matches[2]

		// Read file contents
		filePath := filepath.Join(f.Dir, file.Name())
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		// Split content into separate statements
		statements := strings.Split(string(content), ";")
		var queries []string
		for _, stmt := range statements {
			if strings.TrimSpace(stmt) != "" {
				queries = append(queries, stmt)
			}
		}

		// Create or update migration
		migration, ok := migMap[id]
		if !ok {
			migration = &Migration{
				Id: id,
			}
			migMap[id] = migration
			migrations = append(migrations, migration)
		}

		// Set direction
		switch direction {
		case "up":
			migration.Up = queries
		case "down":
			migration.Down = queries
		}
	}

	// Sort migrations by ID
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Id < migrations[j].Id
	})

	return migrations, nil
}

// StringMigrationSource implements MigrationSource for migrations stored as strings
type StringMigrationSource struct {
	Migrations []*Migration
}

// FindMigrations returns the defined migrations
func (s StringMigrationSource) FindMigrations() ([]*Migration, error) {
	return s.Migrations, nil
}
