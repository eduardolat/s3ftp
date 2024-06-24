package config

import (
	"regexp"
	"time"
)

func validateEnv(env *Env) {
	validateSftpUsers(env)
	validateSyncInterval(env)
	validateSyncMode(env)
}

func validateSftpUsers(env *Env) {
	re := regexp.MustCompile(`^[a-zA-Z0-9_\-:@/.,]+$`)
	if !re.MatchString(*env.SFTP_USERS) {
		logFatalError(
			"SFTP_USERS contains invalid characters",
			"value", *env.SFTP_USERS,
		)
	}
}

func validateSyncInterval(env *Env) {
	_, err := time.ParseDuration(*env.SYNC_INTERVAL)
	if err != nil {
		logFatalError(
			"SYNC_INTERVAL is invalid",
			"value", *env.SYNC_INTERVAL,
		)
	}
}

func validateSyncMode(env *Env) {
	if *env.SYNC_MODE != "sync" && *env.SYNC_MODE != "bisync" {
		logFatalError(
			"SYNC_MODE is invalid, must be 'sync' or 'bisync'",
			"value", *env.SYNC_MODE,
		)
	}
}
