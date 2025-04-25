package utils

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"strings"
	"time"
)

var ansiPattern = regexp.MustCompile("\x1b\\[[0-9;]*m")

// StripAnsi removes ANSI color codes from text
func StripAnsi(text string) string {
	return ansiPattern.ReplaceAllString(text, "")
}

// GetUsername returns current username
func GetUsername() string {
	u, err := user.Current()
	if err != nil {
		return "unknown"
	}
	return u.Username
}

// GetPlatform returns the current platform
func GetPlatform() string {
	return os.Getenv("GOOS")
}

// GetShellName returns the shell being used
func GetShellName() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return "unknown"
	}

	parts := strings.Split(shell, "/")
	return parts[len(parts)-1]
}

// GetTerminalWidth returns the width of the terminal
func GetTerminalWidth() int {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return 80 // default width
	}

	parts := strings.Split(string(out), " ")
	if len(parts) < 2 {
		return 80
	}

	width := 0
	fmt.Sscanf(parts[1], "%d", &width)
	if width <= 0 {
		return 80
	}

	return width
}

// RandomString generates a random string of specified length
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

	rand.Seed(time.Now().UnixNano())

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}

// GetErrorMessageFace returns a random error message face
func GetErrorMessageFace() string {
	faces := []string{
		"⟨ ×︵× ⟩",
		"⟨ ⊖︵⊖ ⟩",
		"⟨ ⊗︵⊗ ⟩",
		"【 ◑︵◐ 】",
		"⦗ ⊘︵⊘ ⦘",
		"[⊗ _ ⊗]",
		"⦅ •︵• ⦆",
		"⟦ ⊖﹏⊖ ⟧",
		"⟨ ⊝︵⊝ ⟩",
		"【 ⊛︵⊛ 】",
		"⦗ ⊕︵⊕ ⦘",
		"⟦ ⊗﹏⊗ ⟧",
		"⦅ ◉︵◉ ⦆",
		"⟨ ⊜﹏⊜ ⟩",
		"【 ⊛﹏⊛ 】",
	}

	rand.Seed(time.Now().UnixNano())
	return faces[rand.Intn(len(faces))]
}
