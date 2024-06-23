package main

import (
	"errors"
	"fmt"
	"s3ftp/internal/config"
	"s3ftp/internal/sftp"
	"strings"
)

func setupSFTP(env *config.Env) error {
	err := sftp.GenerateSSHKeys()
	if err != nil {
		return fmt.Errorf("generate-ssh-keys: %w", err)
	}

	err = sftp.WriteInitialSSHConfig()
	if err != nil {
		return fmt.Errorf("write-initial-ssh-config: %w", err)
	}

	users := strings.Split(*env.SFTP_USERS, ",")
	for _, user := range users {
		userSegments := strings.Split(user, ":")
		if len(userSegments) != 2 && len(userSegments) != 3 {
			return errors.New("invalid SFTP_USERS format")
		}

		username := userSegments[0]
		password := userSegments[1]

		readOnlyMode := false
		if len(userSegments) > 2 {
			readOnlyMode = userSegments[2] == "ro"
		}

		err = sftp.AddUser(username, password, readOnlyMode)
		if err != nil {
			return fmt.Errorf("add-user(%s): %w", username, err)
		}
	}

	sftp.StartSSHD()

	return nil
}
