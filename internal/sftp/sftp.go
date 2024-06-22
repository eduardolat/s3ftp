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
	userDir := "/home/" + user
	uploadsDir := userDir + "/uploads"

	// Create the user
	_, err := exec.Command("adduser", "-D", "-h", userDir, "-s", "/sbin/nologin", user).Output()
	if err != nil {
		return fmt.Errorf("error adding user: %w", err)
	}

	// Set the user's password
	cmd := fmt.Sprintf(`echo "%s:%s" | chpasswd`, user, password)
	_, err = exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return fmt.Errorf("error setting user password: %w", err)
	}

	// Set the correct permissions
	_, err = exec.Command("chown", "root:root", userDir).Output()
	if err != nil {
		return fmt.Errorf("error setting user directory ownership: %w", err)
	}
	_, err = exec.Command("chmod", "755", userDir).Output()
	if err != nil {
		return fmt.Errorf("error setting user directory permissions: %w", err)
	}
	_, err = exec.Command("mkdir", uploadsDir).Output()
	if err != nil {
		return fmt.Errorf("error creating user uploads directory: %w", err)
	}
	_, err = exec.Command("chown", user+":"+user, uploadsDir).Output()
	if err != nil {
		return fmt.Errorf("error setting user uploads directory ownership: %w", err)
	}
	_, err = exec.Command("chmod", "755", uploadsDir).Output()
	if err != nil {
		return fmt.Errorf("error setting user uploads directory permissions: %w", err)
	}

	// Add the user to the sshd_config file
	template := sshUserTemplateRW
	if isReadOnly {
		template = sshUserTemplateRO
	}
	sshdUserConfig := fmt.Sprintf(template, user, userDir)
	f, err := os.OpenFile("/etc/ssh/sshd_config", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening sshd_config: %w", err)
	}
	defer f.Close()
	_, err = f.WriteString(sshdUserConfig)
	if err != nil {
		return fmt.Errorf("error writing to sshd_config: %w", err)
	}

	slog.Info("user added", "user", user)
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
