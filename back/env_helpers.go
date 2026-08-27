package main

import "os"

func isProduction() bool {
	return os.Getenv("RAILWAY_ENVIRONMENT") != "" || os.Getenv("APP_ENV") == "production"
}
