package main

import (
	"os"
	"path/filepath"
)

// uploadsRoot returns the persistent directory used for uploaded files.
// Locally it keeps the existing ../uploads layout; on Railway set UPLOADS_DIR
// to a mounted persistent-volume path such as /app/uploads.
func uploadsRoot() string {
	if dir := os.Getenv("UPLOADS_DIR"); dir != "" {
		return dir
	}
	return filepath.Join("..", "uploads")
}
