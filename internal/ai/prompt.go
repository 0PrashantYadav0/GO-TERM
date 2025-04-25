package ai

import (
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/fatih/color"
)

// FormatPrompt returns a beautifully formatted command prompt
func FormatPrompt() string {
	// Check if colors are supported and enabled
	color.NoColor = os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb"

	// Get username
	u, err := user.Current()
	username := "user"
	if err == nil {
		username = u.Username
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

	// Format directories with different colors for visual depth
	dirs := strings.Split(currentDir, string(os.PathSeparator))
	lastDir := dirs[len(dirs)-1]
	if lastDir == "" && len(dirs) > 1 {
		lastDir = dirs[len(dirs)-2]
	}

	// Build a more colorful path with gradient-like effect
	coloredPath := ""
	if currentDir == "~" || currentDir == "/" {
		coloredPath = color.New(color.FgHiBlue).Sprint(currentDir)
	} else {
		// Format path components with different intensity colors
		pathParts := strings.Split(currentDir, string(os.PathSeparator))
		for i, part := range pathParts {
			if part == "" {
				continue
			}

			if i == 0 && part == "~" {
				coloredPath += color.New(color.FgBlue).Sprint("~")
			} else if i == len(pathParts)-1 && part != "" { // Last directory
				coloredPath += color.New(color.FgHiGreen, color.Bold).Sprint(part)
			} else {
				coloredPath += color.New(color.FgCyan).Sprint(part)
			}

			if i < len(pathParts)-1 {
				coloredPath += color.New(color.FgHiBlack).Sprint("/")
			}
		}
	}

	// Get current time for the prompt
	timeStr := color.New(color.FgHiBlack).Sprint(time.Now().Format("15:04:05"))

	// Create a decorative username display
	userDisplay := color.New(color.FgMagenta, color.Bold).Sprint(username)

	// Beautiful prompt symbols with compatibility for different terminals
	arrow := color.New(color.FgHiCyan).Sprint("❯")

	// Use a simpler arrow if terminal might have issues with Unicode
	if os.Getenv("TERM") == "xterm" || os.Getenv("TERM") == "" {
		arrow = color.New(color.FgHiCyan).Sprint(">")
	}

	// Assemble the final prompt with nice spacing and symbols
	return timeStr + " " + userDisplay + " " + coloredPath + " " + arrow + " "
}

// Force enable colors (call this from main if needed)
func EnableColors() {
	// Force enable colors
	color.NoColor = false
}
