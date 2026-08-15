package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func ConnectDB() {

	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL != "" {

		db, err := sql.Open("postgres", databaseURL)

		if err != nil {
			log.Fatal("DATABASE CONNECTION ERROR:", err)
		}

		if err := db.Ping(); err != nil {
			log.Fatal("DATABASE PING ERROR:", err)
		}

		DB = db

		log.Println("PostgreSQL connected via DATABASE_URL")

		return
	}

	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := os.Getenv("DB_PASSWORD")
	dbname := getEnv("DB_NAME", "clinic")
	sslmode := getEnv("DB_SSLMODE", "disable")

	connStr :=
		"host=" + host +
			" port=" + port +
			" user=" + user +
			" password=" + password +
			" dbname=" + dbname +
			" sslmode=" + sslmode

	db, err := sql.Open("postgres", connStr)

	if err != nil {
		log.Fatal("DATABASE CONNECTION ERROR:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("DATABASE PING ERROR:", err)
	}

	DB = db

	log.Println("PostgreSQL connected locally")
}

func getEnv(key string, defaultValue string) string {

	value := os.Getenv(key)

	if value == "" {
		return defaultValue
	}

	return value
}
