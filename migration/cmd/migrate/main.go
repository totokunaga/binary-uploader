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

func main() {
	host := "localhost"
	port := "13306"
	user := "root"
	password := ""
	dbName := "fs_store"

	if os.Getenv("DB_HOST") != "" {
		host = os.Getenv("DB_HOST")
	}
	if os.Getenv("DB_PORT") != "" {
		port = os.Getenv("DB_PORT")
	}
	if os.Getenv("DB_USER") != "" {
		user = os.Getenv("DB_USER")
	}
	if os.Getenv("DB_PASSWORD") != "" {
		password = os.Getenv("DB_PASSWORD")
	}
	if os.Getenv("DB_NAME") != "" {
		dbName = os.Getenv("DB_NAME")
	}

	// DSN without database for initial connection
	baseDSNWithoutDB := fmt.Sprintf("%s:%s@tcp(%s:%s)/",
		user, password, host, port)

	// DSN with database for migrations
	baseDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?multiStatements=true",
		user, password, host, port, dbName)

	// First, connect without database to create it if needed
	db, err := sql.Open("mysql", baseDSNWithoutDB)
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}
	defer db.Close()

	// Create database if it doesn't exist
	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS " + dbName)
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	log.Printf("Ensured database %s exists", dbName)

	// Add mysql:// prefix for golang-migrate
	defaultDSN := "mysql://" + baseDSN

	// Define command line flags
	dsn := flag.String("dsn", defaultDSN, "Database DSN")
	migrationPath := flag.String("path", "file://db/migrations", "Path to migration files")
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

	switch *command {
	case "up":
		if *steps > 0 {
			err = m.Steps(*steps)
		} else {
			err = m.Up()
		}
	case "down":
		if *steps > 0 {
			err = m.Steps(-(*steps))
		} else {
			err = m.Down()
		}
	case "version":
		version, dirty, vErr := m.Version()
		if vErr != nil {
			if vErr == migrate.ErrNilVersion {
				fmt.Println("No migration has been applied yet")
				return
			}
			log.Fatalf("Error getting migration version: %v", vErr)
		}
		fmt.Printf("Current migration version: %v, Dirty: %v\n", version, dirty)
		return
	case "force":
		if *version < 0 {
			log.Fatalf("Version must be specified for force command")
		}
		err = m.Force(*version)
	default:
		flag.Usage()
		os.Exit(1)
	}

	if err != nil {
		if err == migrate.ErrNoChange {
			fmt.Println("No migration needed")
			return
		}
		log.Fatalf("Migration error: %v", err)
	}

	fmt.Printf("Migration '%s' completed successfully\n", *command)
}
