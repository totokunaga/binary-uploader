package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	DefaultHost     = "localhost"
	DefaultPort     = "33306"
	DefaultUser     = "superuser"
	DefaultPassword = "superpass"
	DefaultDBName   = "fs_store"

	MigrationDirPath = "file://db/migrations"
)

func main() {
	dbConfig := parseEnvVars()

	// DSN without database for initial connection
	baseDSNWithoutDB := fmt.Sprintf("%s:%s@tcp(%s:%s)/",
		dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port)

	// DSN with database for migrations
	baseDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?multiStatements=true",
		dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.DBName)

	// Connect with a non-database connection URL to create the database
	db, err := sql.Open("mysql", baseDSNWithoutDB)
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}
	defer db.Close()

	// Create database if it doesn't exist
	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS " + dbConfig.DBName)
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	log.Printf("Ensured database %s exists", dbConfig.DBName)

	// Add mysql:// prefix for golang-migrate
	defaultDSN := "mysql://" + baseDSN

	// Define command line flags
	dsn := flag.String("dsn", defaultDSN, "Database DSN")
	migrationPath := flag.String("path", MigrationDirPath, "Path to migration files")
	command := flag.String("command", "up", "Migration command (up, down, version, force)")
	steps := flag.Int("steps", 0, "Number of migrations to apply (0 = all)")
	version := flag.Int("version", -1, "Target version for force command")

	flag.Parse()

	// Create a new migrate instance
	m, err := migrate.New(*migrationPath, *dsn)
	if err != nil {
		log.Fatalf("Migration initialization error: %v", err)
	}

	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			log.Printf("Error closing source: %v", srcErr)
		}
		if dbErr != nil {
			log.Printf("Error closing database: %v", dbErr)
		}
	}()

	// Execute the migration command
	execErr := execMigrationCommand(m, *command, *steps, *version)

	if execErr != nil {
		if execErr == migrate.ErrNoChange {
			fmt.Println("No migration needed")
			return
		}
		log.Fatalf("Migration error: %v", execErr)
	}

	fmt.Printf("Migration '%s' completed successfully\n", *command)
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// parseEnvVars parses environment variables for database configuration
func parseEnvVars() DBConfig {
	config := DBConfig{
		Host:     DefaultHost,
		Port:     DefaultPort,
		User:     DefaultUser,
		Password: DefaultPassword,
		DBName:   DefaultDBName,
	}

	if os.Getenv("DB_HOST") != "" {
		config.Host = os.Getenv("DB_HOST")
	}
	if os.Getenv("DB_PORT") != "" {
		config.Port = os.Getenv("DB_PORT")
	}
	if os.Getenv("DB_USER") != "" {
		config.User = os.Getenv("DB_USER")
	}
	if os.Getenv("DB_PASSWORD") != "" {
		config.Password = os.Getenv("DB_PASSWORD")
	}
	if os.Getenv("DB_NAME") != "" {
		config.DBName = os.Getenv("DB_NAME")
	}

	return config
}

// execMigrationCommand executes the migration command
func execMigrationCommand(m *migrate.Migrate, command string, steps int, version int) error {
	switch command {
	case "up":
		if steps > 0 {
			return m.Steps(steps)
		} else {
			return m.Up()
		}
	case "down":
		if steps > 0 {
			return m.Steps(-steps)
		} else {
			return m.Down()
		}
	case "version":
		version, dirty, vErr := m.Version()
		if vErr != nil {
			if vErr == migrate.ErrNilVersion {
				return fmt.Errorf("no migration has been applied yet")
			}
			return fmt.Errorf("error getting migration version: %v", vErr)
		}
		fmt.Printf("Current migration version: %v, Dirty: %v\n", version, dirty)
		return nil
	case "force":
		if version < 0 {
			return fmt.Errorf("version must be specified for force command")
		}
		return m.Force(version)
	default:
		return fmt.Errorf("invalid command: %s", command)
	}
}
