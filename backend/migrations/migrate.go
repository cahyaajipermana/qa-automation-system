package migrations

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

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
	files, err := ioutil.ReadDir(migrationsDir)
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
		content, err := ioutil.ReadFile(filepath.Join(migrationsDir, file.Name()))
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %v", file.Name(), err)
		}

		// Start transaction
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to start transaction: %v", err)
		}

		// Execute migration
		_, err = tx.Exec(string(content))
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
	content, err := ioutil.ReadFile(filepath.Join(migrationsDir, downFile))
	if err != nil {
		return fmt.Errorf("failed to read rollback file %s: %v", downFile, err)
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}

	// Execute rollback
	_, err = tx.Exec(string(content))
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