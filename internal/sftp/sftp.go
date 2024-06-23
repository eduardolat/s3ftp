package sftp

import (
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
)

//go:embed sshd_config
var sshdConfig string

// sshUserTemplateRW is a template for adding a user to the sshd_config file in read-write mode
const sshUserTemplateRW = `
Match User %s
  ChrootDirectory %s
  ForceCommand internal-sftp
  AllowTcpForwarding no
  X11Forwarding no
`

// sshUserTemplateRO is a template for adding a user to the sshd_config file in read-only mode
const sshUserTemplateRO = `
Match User %s
	ChrootDirectory %s
	ForceCommand internal-sftp -R
	AllowTcpForwarding no
	X11Forwarding no
`

// WriteInitialSSHConfig writes the initial sshd_config file to /etc/ssh/sshd_config
func WriteInitialSSHConfig() error {
	sshdDir := "/etc/ssh"
	sshdPath := "/etc/ssh/sshd_config"

	// Create the directory if it doesn't exist
	err := os.MkdirAll(sshdDir, 0755)
	if err != nil {
		return fmt.Errorf("error creating sshd directory: %w", err)
	}

	// Delete the file if it exists
	err = os.Remove(sshdPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error deleting sshd_config: %w", err)
	}

	// Write the file
	f, err := os.Create(sshdPath)
	if err != nil {
		return fmt.Errorf("error creating sshd_config: %w", err)
	}
	defer f.Close()
	_, err = f.WriteString(sshdConfig)
	if err != nil {
		return fmt.Errorf("error writing to sshd_config: %w", err)
	}

	slog.Info("initial sshd_config written")
	return nil
}

// AddUser adds a user to the system, sets the necessary permissions, and adds the user
// to the sshd_config file
func AddUser(user, password string, isReadOnly bool) error {
	chrootDir := fmt.Sprintf("/home/%s", user)
	userDir := fmt.Sprintf("/home/%s/%s", user, user)

	commands := []command{
		{
			name: "create chroot dir",
			cmd:  fmt.Sprintf("mkdir -p %s", chrootDir),
		},
		{
			name: "create user dir",
			cmd:  fmt.Sprintf("mkdir -p %s", userDir),
		},
		{
			name: "add user",
			cmd:  fmt.Sprintf("adduser -D -h %s -s /sbin/nologin %s", chrootDir, user),
		},
		{
			name: "set user password",
			cmd:  fmt.Sprintf(`echo "%s:%s" | chpasswd`, user, password),
		},
		{
			name: "set chroot dir ownership",
			cmd:  fmt.Sprintf("chown root:root %s", chrootDir),
		},
		{
			name: "set chroot dir permissions",
			cmd:  fmt.Sprintf("chmod 755 %s", chrootDir),
		},
		{
			name: "set user dir ownership",
			cmd:  fmt.Sprintf("chown %s:%s %s", user, user, userDir),
		},
		{
			name: "set user dir permissions",
			cmd:  fmt.Sprintf("chmod 700 %s", userDir),
		},
	}

	for _, cmd := range commands {
		_, err := execNamedCMD(cmd)
		if err != nil {
			return err
		}
	}

	// Add the user to the sshd_config file
	template := sshUserTemplateRW
	if isReadOnly {
		template = sshUserTemplateRO
	}
	sshdUserConfig := fmt.Sprintf(template, user, chrootDir)
	f, err := os.OpenFile("/etc/ssh/sshd_config", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening sshd_config: %w", err)
	}
	defer f.Close()
	_, err = f.WriteString(sshdUserConfig)
	if err != nil {
		return fmt.Errorf("error writing to sshd_config: %w", err)
	}

	if isReadOnly {
		slog.Info(fmt.Sprintf("user %s added as ro user", user))
	} else {
		slog.Info(fmt.Sprintf("user %s added as rw user", user))
	}

	return nil
}

// GenerateSSHKeys generates the necessary keys for the sftp server
func GenerateSSHKeys() error {
	_, err := exec.Command("ssh-keygen", "-A").Output()
	if err != nil {
		return fmt.Errorf("error generating ssh keys: %w", err)
	}

	slog.Info("ssh keys generated")
	return nil
}

// StartSSHD starts the sshd service
func StartSSHD() error {
	cmd := exec.Command("/usr/sbin/sshd", "-D")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting sshd: %w", err)
	}
	slog.Info("sshd started")

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("error waiting for sshd to finish: %w", err)
	}
	slog.Info("sshd finished")

	return nil
}
