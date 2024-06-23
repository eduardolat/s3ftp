package sftp

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type user struct {
	Username string
	HomeDir  string
	Group    string
}

func getUsers() ([]user, error) {
	file, err := os.Open("/etc/passwd")
	if err != nil {
		return nil, fmt.Errorf("error opening /etc/passwd: %w", err)
	}
	defer file.Close()

	users := []user{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ":")
		if len(fields) < 7 {
			continue
		}
		username := fields[0]
		homeDir := fields[5]
		groupID := fields[3]
		group, err := getGroupName(groupID)
		if err != nil {
			return nil, fmt.Errorf("error getting group name for %s: %w", username, err)
		}
		users = append(users, user{Username: username, HomeDir: homeDir, Group: group})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading /etc/passwd: %w", err)
	}
	return users, nil
}

func getGroupName(gid string) (string, error) {
	file, err := os.Open("/etc/group")
	if err != nil {
		return "", fmt.Errorf("error opening /etc/group: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, ":")
		if len(fields) < 3 {
			continue
		}
		if fields[2] == gid {
			return fields[0], nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading /etc/group: %w", err)
	}
	return "", fmt.Errorf("group with GID %s not found", gid)
}
