package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"qa-automation-system/backend/migrations"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Parse command line flags
	action := flag.String("action", "up", "Migration action (up or down)")
	dsn := flag.String("dsn", "", "Database connection string")
	flag.Parse()

	// If DSN is not provided via flag, try to get it from environment variables
	if *dsn == "" {
		dbHost := os.Getenv("DB_HOST")
		dbPort := os.Getenv("DB_PORT")
		dbUser := os.Getenv("DB_USER")
		dbPass := os.Getenv("DB_PASSWORD")
		dbName := os.Getenv("DB_NAME")

		if dbHost == "" || dbPort == "" || dbUser == "" || dbPass == "" || dbName == "" {
			log.Fatal("Database connection details not found. Please provide DSN via -dsn flag or set environment variables (DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME)")
		}

		*dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)
	}

	// Connect to database
	db, err := sql.Open("mysql", *dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test the connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Get the absolute path to the migrations directory
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}
	migrationsDir := filepath.Join(currentDir, "migrations")

	// Check if migrations directory exists
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		log.Fatalf("Migrations directory not found: %s", migrationsDir)
	}

	// Run migrations based on action
	switch *action {
	case "up":
		if err := migrations.RunMigrations(db, migrationsDir); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		fmt.Println("Migrations completed successfully")
	case "down":
		if err := migrations.RollbackMigrations(db, migrationsDir); err != nil {
			log.Fatalf("Failed to rollback migrations: %v", err)
		}
		fmt.Println("Rollback completed successfully")
	default:
		log.Fatalf("Invalid action: %s. Use 'up' or 'down'", *action)
	}
} 