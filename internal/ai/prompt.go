package ai

import (
	"os"
	"os/user"
	"strings"

	"github.com/fatih/color"
)

// FormatPrompt returns a formatted command prompt
func FormatPrompt() string {
	// Get username
	u, err := user.Current()
	username := "user"
	if err == nil {
		username = u.Username
	}

	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}

	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		currentDir = "/"
	}

	// Replace home directory with ~
	homeDir, err := os.UserHomeDir()
	if err == nil && strings.HasPrefix(currentDir, homeDir) {
		currentDir = "~" + currentDir[len(homeDir):]
	}

	// Format the last directory with a different color
	dirs := strings.Split(currentDir, string(os.PathSeparator))
	lastDir := dirs[len(dirs)-1]
	if lastDir == "" && len(dirs) > 1 {
		lastDir = dirs[len(dirs)-2]
	}

	pathDisplay := strings.ReplaceAll(currentDir, lastDir, color.GreenString(lastDir))

	return color.BlueString(username) + "@" + color.WhiteString(hostname) + " " +
		strings.ReplaceAll(pathDisplay, "~", color.BlueString("~")) + "> "
}
