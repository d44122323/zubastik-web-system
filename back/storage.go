package main

import (
	"os"
	"path/filepath"
)

func uploadsRoot() string {
	if dir := os.Getenv("UPLOADS_DIR"); dir != "" {
		return dir
	}
	return filepath.Join("..", "uploads")
}
