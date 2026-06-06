package database

import (
	"fmt"
	"log"
	"time"
	"github.com/delaquash/carezo/configs"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var DB *sqlx.DB

func ConnectPostgres(cfg *configs.Config) (*sqlx.DB, error) {
	// Build connection string (DSN - Data Source Name)
	// Format: postgres://username:password@host:port/database?options
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	// Open connection to database and sqlx.Connect also pings the db to verif
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(25)  
	db.SetMaxIdleConns(5)  
	db.SetConnMaxIdleTime(5 * time.Minute)


	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Connected to PostgreSQL successfully")
	
	// Store in global variable for easy access
	DB = db
	
	return db, nil
}

// ClosePostgres closes database connection gracefully
func ClosePostgres() error {
	if DB != nil {
		log.Println("Closing PostgreSQL connection...")
		return DB.Close()
	}
	return nil
}

