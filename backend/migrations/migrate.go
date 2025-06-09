package migrations

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// migrationsDir is the directory containing migration files
var migrationsDir = filepath.Join("migrations")

// Migration represents a database migration
type Migration struct {
	Version int
	Name    string
	Up      string
	Down    string
}

// RunMigrations executes all pending migrations
func RunMigrations(db *sql.DB, migrationsDir string) error {
	// Create migrations table if it doesn't exist
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id INT AUTO_INCREMENT PRIMARY KEY,
			version INT NOT NULL,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %v", err)
	}

	// Get all migration files
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %v", err)
	}

	// Sort files by version
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	// Get applied migrations
	rows, err := db.Query("SELECT version FROM migrations ORDER BY version")
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %v", err)
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("failed to scan migration version: %v", err)
		}
		applied[version] = true
	}

	// Apply pending migrations
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".up.sql") {
			continue
		}

		version := getVersionFromFilename(file.Name())
		if applied[version] {
			continue
		}

		// Read migration file
		content, err := readMigrationFile(filepath.Join(migrationsDir, file.Name()))
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %v", file.Name(), err)
		}

		// Start transaction
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to start transaction: %v", err)
		}

		// Execute migration
		_, err = tx.Exec(content)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %v", file.Name(), err)
		}

		// Record migration
		_, err = tx.Exec("INSERT INTO migrations (version, name) VALUES (?, ?)", version, file.Name())
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %v", file.Name(), err)
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %v", file.Name(), err)
		}

		log.Printf("Applied migration: %s", file.Name())
	}

	return nil
}

// RollbackMigrations rolls back the last migration
func RollbackMigrations(db *sql.DB, migrationsDir string) error {
	// Get the last applied migration
	var version int
	var name string
	err := db.QueryRow("SELECT version, name FROM migrations ORDER BY version DESC LIMIT 1").Scan(&version, &name)
	if err != nil {
		return fmt.Errorf("no migrations to rollback: %v", err)
	}

	// Read down migration file
	downFile := strings.Replace(name, ".up.sql", ".down.sql", 1)
	content, err := readMigrationFile(filepath.Join(migrationsDir, downFile))
	if err != nil {
		return fmt.Errorf("failed to read rollback file %s: %v", downFile, err)
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}

	// Execute rollback
	_, err = tx.Exec(content)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to execute rollback %s: %v", downFile, err)
	}

	// Remove migration record
	_, err = tx.Exec("DELETE FROM migrations WHERE version = ?", version)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to remove migration record: %v", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback: %v", err)
	}

	log.Printf("Rolled back migration: %s", name)
	return nil
}

// getVersionFromFilename extracts the version number from a migration filename
func getVersionFromFilename(filename string) int {
	var version int
	fmt.Sscanf(filename, "%d_", &version)
	return version
}

// readMigrationFile reads the content of a migration file
func readMigrationFile(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("failed to open migration file: %v", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read migration file: %v", err)
	}

	return string(content), nil
}

// getMigrationVersion extracts the version number from a migration filename
func getMigrationVersion(filename string) int {
	var version int
	_, err := fmt.Sscanf(filename, "%d_", &version)
	if err != nil {
		// If we can't parse the version, return 0 to indicate an invalid migration
		return 0
	}
	return version
}

// MigrateUp applies all pending migrations
func MigrateUp(db *sql.DB) error {
	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	// Get all migration files
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("failed to read migrations directory: %v (rollback error: %v)", err, rbErr)
		}
		return fmt.Errorf("failed to read migrations directory: %v", err)
	}

	// Sort files by version
	var migrations []struct {
		name    string
		version int
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".up.sql") {
			version := getMigrationVersion(file.Name())
			if version == 0 {
				if rbErr := tx.Rollback(); rbErr != nil {
					return fmt.Errorf("invalid migration filename %s (rollback error: %v)", file.Name(), rbErr)
				}
				return fmt.Errorf("invalid migration filename %s", file.Name())
			}
			migrations = append(migrations, struct {
				name    string
				version int
			}{file.Name(), version})
		}
	}
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})

	// Apply migrations
	for _, migration := range migrations {
		// Read migration file
		content, err := readMigrationFile(filepath.Join(migrationsDir, migration.name))
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				return fmt.Errorf("failed to read migration file %s: %v (rollback error: %v)", migration.name, err, rbErr)
			}
			return fmt.Errorf("failed to read migration file %s: %v", migration.name, err)
		}

		// Execute migration
		_, err = tx.Exec(content)
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				return fmt.Errorf("failed to execute migration %s: %v (rollback error: %v)", migration.name, err, rbErr)
			}
			return fmt.Errorf("failed to execute migration %s: %v", migration.name, err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("failed to commit transaction: %v (rollback error: %v)", err, rbErr)
		}
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// MigrateDown rolls back the last migration
func MigrateDown(db *sql.DB) error {
	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	// Get all migration files
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("failed to read migrations directory: %v (rollback error: %v)", err, rbErr)
		}
		return fmt.Errorf("failed to read migrations directory: %v", err)
	}

	// Find the latest migration
	var latestMigration string
	var latestVersion int
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".up.sql") {
			version := getMigrationVersion(file.Name())
			if version > latestVersion {
				latestVersion = version
				latestMigration = file.Name()
			}
		}
	}

	if latestMigration == "" {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("no migrations found (rollback error: %v)", rbErr)
		}
		return fmt.Errorf("no migrations found")
	}

	// Read down migration file
	downFile := strings.Replace(latestMigration, ".up.sql", ".down.sql", 1)
	content, err := readMigrationFile(filepath.Join(migrationsDir, downFile))
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("failed to read rollback file %s: %v (rollback error: %v)", downFile, err, rbErr)
		}
		return fmt.Errorf("failed to read rollback file %s: %v", downFile, err)
	}

	// Execute rollback
	_, err = tx.Exec(content)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("failed to execute rollback %s: %v (rollback error: %v)", downFile, err, rbErr)
		}
		return fmt.Errorf("failed to execute rollback %s: %v", downFile, err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("failed to commit transaction: %v (rollback error: %v)", err, rbErr)
		}
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
} 