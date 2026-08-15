package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {

	loadEnv()

	InitAI()

	cfg := LoadConfig()

	ConnectDB()

	defer DB.Close()

	mux := SetupRouter(cfg)

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	addr := ":" + port

	fmt.Println("SERVER RUNNING ON PORT:", port)

	log.Fatal(
		http.ListenAndServe(addr, mux),
	)
}

func loadEnv() {

	if err := godotenv.Load(".env"); err == nil {

		log.Println("ENV FILE LOADED: .env")

		return
	}

	if err := godotenv.Load("back/.env"); err == nil {

		log.Println("ENV FILE LOADED: back/.env")

		return
	}

	log.Println("ENV FILE NOT FOUND - USING ENVIRONMENT VARIABLES")
}
