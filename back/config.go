package main

import "os"

type Config struct {
	TelegramBotToken string
	TelegramChatID   string
	EmailUser        string
	EmailPassword    string
	BitrixWebhook    string
}

func LoadConfig() Config {
	return Config{
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChatID:   os.Getenv("TELEGRAM_CHAT_ID"),

		EmailUser:     os.Getenv("EMAIL_USER"),
		EmailPassword: os.Getenv("EMAIL_PASSWORD"),

		BitrixWebhook: os.Getenv("BITRIX_WEBHOOK"),
	}
}
