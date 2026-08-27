package main

import "os"

type Config struct {
	// Patient bot
	TelegramBotToken    string
	TelegramBotUsername string

	// Staff bot (doctor/admin)
	TelegramStaffBotToken    string
	TelegramStaffBotUsername string
	TelegramStaffChatID      string

	EmailUser     string
	EmailPassword string
	BitrixWebhook string
}

func LoadConfig() Config {
	return Config{
		TelegramBotToken:         os.Getenv("TELEGRAM_PATIENT_BOT_TOKEN"),
		TelegramBotUsername:      os.Getenv("TELEGRAM_PATIENT_BOT_USERNAME"),
		TelegramStaffBotToken:    os.Getenv("TELEGRAM_STAFF_BOT_TOKEN"),
		TelegramStaffBotUsername: os.Getenv("TELEGRAM_STAFF_BOT_USERNAME"),
		TelegramStaffChatID:      os.Getenv("TELEGRAM_STAFF_CHAT_ID"),
		EmailUser:                os.Getenv("EMAIL_USER"),
		EmailPassword:            os.Getenv("EMAIL_PASSWORD"),
		BitrixWebhook:            os.Getenv("BITRIX_WEBHOOK"),
	}
}
