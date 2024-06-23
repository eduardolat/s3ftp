package sftp

import (
	"fmt"
	"os/exec"
)

type command struct {
	name string
	cmd  string
}

func execNamedCMD(cmd command) ([]byte, error) {
	b, err := exec.Command("sh", "-c", cmd.cmd).Output()
	if err != nil {
		return nil, fmt.Errorf("error %s: %w", cmd.name, err)
	}
	return b, nil
}
