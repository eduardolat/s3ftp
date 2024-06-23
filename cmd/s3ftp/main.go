package main

import (
	"log/slog"
	"os"
	"s3ftp/internal/config"
	"s3ftp/internal/sftp"
)

func main() {
	slog.Info("Starting s3ftp...")
	env := config.GetEnv()

	if err := sftp.ResetSFTP(); err != nil {
		slog.Error("error resetting SFTP", "error", err)
		os.Exit(1)
	}

	if err := sftp.SetupSFTP(env); err != nil {
		slog.Error("error setting up SFTP", "error", err)
		os.Exit(1)
	}

	sftp.StartSSHD()
}
