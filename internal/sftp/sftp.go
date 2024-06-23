package sftp

import (
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"s3ftp/internal/config"
	"strings"
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

// usersGroup is the group that all users belong to
const usersGroup = "s3ftp-users"

// writeInitialSSHConfig writes the initial sshd_config file to /etc/ssh/sshd_config
func writeInitialSSHConfig() error {
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

// createUsersGroup creates a group with the given name
func createUsersGroup() error {
	cmd := fmt.Sprintf("addgroup %s", usersGroup)
	_, err := execNamedCMD(command{name: "create group", cmd: cmd})
	if err != nil {
		return err
	}

	slog.Info(fmt.Sprintf("group %s created", usersGroup))
	return nil
}

// addUser adds a user to the system, sets the necessary permissions, and adds the user
// to the sshd_config file
func addUser(user, password string, isReadOnly bool) error {
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
			cmd: fmt.Sprintf(
				"adduser -D -h %s -s /sbin/nologin -G %s %s", chrootDir, usersGroup, user,
			),
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
			cmd:  fmt.Sprintf("chown %s:%s %s", user, usersGroup, userDir),
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

// generateSSHKeys generates the necessary keys for the sftp server
func generateSSHKeys() error {
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

func SetupSFTP(env *config.Env) error {
	type us struct {
		Username string
		Passwd   string
		ReadOnly bool
	}

	envUsers := strings.Split(*env.SFTP_USERS, ",")
	users := make([]us, len(envUsers))
	usernames := map[string]int{}
	for i, user := range envUsers {
		userSegments := strings.Split(user, ":")
		if len(userSegments) != 2 && len(userSegments) != 3 {
			return errors.New("invalid SFTP_USERS format")
		}

		isReadOnly := false
		if len(userSegments) > 2 {
			isReadOnly = userSegments[2] == "ro"
		}

		users[i] = us{
			Username: userSegments[0],
			Passwd:   userSegments[1],
			ReadOnly: isReadOnly,
		}

		usernames[users[i].Username]++
		if usernames[users[i].Username] > 1 {
			return fmt.Errorf("duplicate username: %s", users[i].Username)
		}
	}

	err := generateSSHKeys()
	if err != nil {
		return fmt.Errorf("generate-ssh-keys: %w", err)
	}

	err = writeInitialSSHConfig()
	if err != nil {
		return fmt.Errorf("write-initial-ssh-config: %w", err)
	}

	err = createUsersGroup()
	if err != nil {
		return fmt.Errorf("create-users-group: %w", err)
	}

	for _, user := range users {
		err = addUser(user.Username, user.Passwd, user.ReadOnly)
		if err != nil {
			return fmt.Errorf("add-user(%s): %w", user.Username, err)
		}
	}

	return nil
}

func ResetSFTP() error {
	users, err := getUsers()
	if err != nil {
		return err
	}
	for _, user := range users {
		if user.Group == usersGroup {
			_, _ = execNamedCMD(command{
				name: "delete user",
				cmd:  fmt.Sprintf("deluser %s", user.Username),
			})
		}
	}

	commands := []command{
		{
			name: "delete ssh keys",
			cmd:  "rm -f /etc/ssh/ssh_host_*",
		},
		{
			name: "delete users group",
			cmd:  fmt.Sprintf("delgroup %s", usersGroup),
		},
		{
			name: "delete sshd_config",
			cmd:  "rm -f /etc/ssh/sshd_config",
		},
	}

	for _, cmd := range commands {
		_, _ = execNamedCMD(cmd)
	}

	slog.Info("s3ftp reset executed")
	return nil
}
