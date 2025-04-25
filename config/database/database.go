package database

import (
	"log"
	"os"
	"sync"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var (
	once    sync.Once
	db_pool *sqlx.DB
)

func initializeDB() {
	DATABASE_URL := os.Getenv("DATABASE_URL")
	var err_db error

	db_pool, err_db = sqlx.Connect("postgres", DATABASE_URL)
	if err_db != nil {
		log.Fatalln("Error connection to the database", err_db)
	}

	if err := db_pool.Ping(); err != nil {
		defer db_pool.Close()
		log.Fatalf("Failed to ping the database: %v", err)
	}

	// Set connection pool settings
	db_pool.SetMaxIdleConns(10)
	db_pool.SetMaxOpenConns(10)
	db_pool.SetConnMaxLifetime(0)
}

func GetDB() *sqlx.DB {
	once.Do(initializeDB)
	return db_pool
}
