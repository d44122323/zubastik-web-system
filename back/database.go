package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func ConnectDB() {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		dbname := os.Getenv("DB_NAME")
		password := os.Getenv("DB_PASSWORD")
		if host == "" || port == "" || user == "" || dbname == "" || password == "" {
			log.Fatal("DB_PASSWORD is required when DATABASE_URL is not set")
		}
		sslmode := envOrDefault("DB_SSLMODE", "disable")
		connStr = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbname, sslmode)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	DB = db
	log.Println("PostgreSQL connected")
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
