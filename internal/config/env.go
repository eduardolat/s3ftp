package config

import (
	"log/slog"

	"github.com/joho/godotenv"
)

type Env struct {
	SFTP_USERS *string

	S3_ACCESS_KEY_ID     *string
	S3_SECRET_ACCESS_KEY *string
	S3_REGION            *string
	S3_ENDPOINT          *string
	S3_BUCKET            *string

	SYNC_CRON *string
}

// GetEnv returns the environment variables.
//
// If there is an error, it will log it and exit the program.
func GetEnv() *Env {
	err := godotenv.Load()
	if err == nil {
		slog.Info("👉 using .env file")
	}

	env := &Env{
		SFTP_USERS: getEnvAsString(getEnvAsStringParams{
			name:       "SFTP_USERS",
			isRequired: true,
		}),

		S3_ACCESS_KEY_ID: getEnvAsString(getEnvAsStringParams{
			name:       "S3_ACCESS_KEY_ID",
			isRequired: true,
		}),
		S3_SECRET_ACCESS_KEY: getEnvAsString(getEnvAsStringParams{
			name:       "S3_SECRET_ACCESS_KEY",
			isRequired: true,
		}),
		S3_REGION: getEnvAsString(getEnvAsStringParams{
			name:       "S3_REGION",
			isRequired: true,
		}),
		S3_ENDPOINT: getEnvAsString(getEnvAsStringParams{
			name:       "S3_ENDPOINT",
			isRequired: true,
		}),
		S3_BUCKET: getEnvAsString(getEnvAsStringParams{
			name:       "S3_BUCKET",
			isRequired: true,
		}),

		SYNC_CRON: getEnvAsString(getEnvAsStringParams{
			name:         "SYNC_CRON",
			defaultValue: newDefaultValue("*/5 * * * *"), // every 5 minutes
		}),
	}

	return env
}
