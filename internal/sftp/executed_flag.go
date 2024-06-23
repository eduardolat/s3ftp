package sftp

import (
	"os"
)

const flagFilePath = "/.s3ftp-executed"

// saveExecuted creates a file to indicate that the setup has been executed.
func saveExecuted() error {
	file, err := os.Create(flagFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString("executed")
	if err != nil {
		return err
	}

	return nil
}

// isExecuted checks if the setup has been executed.
func isExecuted() (bool, error) {
	if _, err := os.Stat(flagFilePath); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}
