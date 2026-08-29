package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()

	if err != nil {
		log.Println("ENV FILE NOT FOUND")
	}

	InitAI()

	cfg := LoadConfig()
	ConnectDB()
	defer DB.Close()
	if err := InitDoctorsTable(); err != nil {
		log.Fatal(err)
	}
	if err := SeedDoctors(); err != nil {
		log.Fatal(err)
	}
	if err := InitDoctorAuthTables(); err != nil {
		log.Fatal(err)
	}
	if err := InitPatientAuthSchema(); err != nil {
		log.Fatal(err)
	}
	if err := InitDoctorBookingSchema(); err != nil {
		log.Fatal(err)
	}
	if err := InitServicesTable(); err != nil {
		log.Fatal(err)
	}
	if err := InitMedicalRecordsSchema(); err != nil {
		log.Fatal(err)
	}
	if err := InitDoctorQuestionsSchema(); err != nil {
		log.Fatal(err)
	}
	if err := InitTelegramSchema(); err != nil {
		log.Fatal(err)
	}
	if err := InitCaptchaSchema(); err != nil {
		log.Fatal(err)
	}
	mux := SetupRouter(cfg)

	mux.(*http.ServeMux).HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := DB.Ping(); err != nil {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	if cfg.TelegramBotToken != "" {
		go telegramPoller(cfg)
		go telegramReminderWorker(cfg)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("SERVER RUNNING on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
