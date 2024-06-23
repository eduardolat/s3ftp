package rclone

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"s3ftp/internal/config"
	"time"
)

const confTemplate = `[s3]
type = s3
provider = Other
access_key_id = %s
secret_access_key = %s
region = %s
endpoint = %s`

// CreateConf creates the rclone configuration file.
func CreateConf(env *config.Env) error {
	path := "/root/.config/rclone/rclone.conf"

	// Delete the file if it exists
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll("/root/.config/rclone", 0755); err != nil {
		return err
	}

	// Create the file
	if _, err := os.Create(path); err != nil {
		return err
	}

	// Write the file
	fileContent := fmt.Sprintf(
		confTemplate,
		*env.S3_ACCESS_KEY_ID,
		*env.S3_SECRET_ACCESS_KEY,
		*env.S3_REGION,
		*env.S3_ENDPOINT,
	)
	if err := os.WriteFile(path, []byte(fileContent), 0644); err != nil {
		return err
	}

	return nil
}

// runBisync runs the rclone bidirectional sync command.
func runBisync(env *config.Env, shouldResync bool) error {
	cmd := fmt.Sprintf("rclone bisync s3:%s/ /home", *env.S3_BUCKET)
	if shouldResync {
		cmd += " --resync"
	}

	_, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	return nil
}

// RunBisyncLoop runs the rclone bidirectional sync command in a loop.
func RunBisyncLoop(env *config.Env) error {
	slog.Info("starting rclone bisync loop...")

	dur, err := time.ParseDuration(*env.SYNC_INTERVAL)
	if err != nil {
		return err
	}

	executions := 0
	for {
		shouldResync := executions == 0
		if err := runBisync(env, shouldResync); err != nil {
			return err
		}

		executions++
		slog.Info(
			"S3 Synced",
			"executions", executions,
			"interval", *env.SYNC_INTERVAL,
			"timestamp", time.Now().Format(time.RFC3339),
			"next_execution", time.Now().Add(dur).Format(time.RFC3339),
			"resync", shouldResync,
		)
		time.Sleep(dur)
	}
}
