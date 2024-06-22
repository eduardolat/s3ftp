package main

import (
	"log/slog"
	"os"
	"s3ftp/internal/config"
)

func main() {
	slog.Info("Starting s3ftp...")
	env := config.GetEnv()

	err := setupSFTP(env)
	if err != nil {
		slog.Error("error setting up SFTP", "error", err)
		os.Exit(1)
	}
}
