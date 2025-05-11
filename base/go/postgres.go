package base

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// PostgresConfig stores the configuration for the PostgreSQL connection.
type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewPostgresClient creates a new PostgreSQL client using GORM.
func NewPostgresClient(config PostgresConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to the 'postgres' database: %v", err)
		return nil, err
	}

	// Check if the database exists
	var dbExists bool
	err = db.Raw("SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = ?)", config.DBName).Scan(&dbExists).Error
	if err != nil {
		log.Fatalf("failed to check if database exists: %v", err)
		return nil, err
	}

	// Create the database if it d	oesn't exist
	if !dbExists {
		err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", config.DBName)).Error
		if err != nil {
			log.Fatalf("failed to create database: %v", err)
			return nil, err
		}
		fmt.Printf("Database %s created!\n", config.DBName)
	}

	// Now connect to the actual database
	dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to the '%s' database: %v", config.DBName, err)
		return nil, err
	}

	fmt.Println("Connected to PostgreSQL!")
	return db, nil
}
