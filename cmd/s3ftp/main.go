package main

import (
	"fmt"
	"log/slog"
	"os"
	"s3ftp/internal/config"
	"s3ftp/internal/rclone"
	"s3ftp/internal/sftp"

	"golang.org/x/sync/errgroup"
)

func main() {
	slog.Info("starting s3ftp...")
	env := config.GetEnv()

	if err := sftp.ResetSFTP(); err != nil {
		slog.Error("error resetting SFTP", "error", err)
		os.Exit(1)
	}

	if err := sftp.SetupSFTP(env); err != nil {
		slog.Error("error setting up SFTP", "error", err)
		os.Exit(1)
	}

	if err := rclone.CreateConf(env); err != nil {
		slog.Error("error creating rclone configuration", "error", err)
		os.Exit(1)
	}

	eg := errgroup.Group{}
	eg.SetLimit(2)

	eg.Go(func() error {
		err := sftp.StartSSHD()
		return fmt.Errorf("SSHD error: %w", err)
	})

	eg.Go(func() error {
		err := rclone.RunBisyncLoop(env)
		return fmt.Errorf("rclone error: %w", err)
	})

	if err := eg.Wait(); err != nil {
		slog.Error("error", "error", err)
		os.Exit(1)
	}
}
