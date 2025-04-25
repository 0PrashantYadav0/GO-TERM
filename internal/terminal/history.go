package terminal

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const HISTORY_LIMIT = 1000

// History manages command history
type History struct {
	filePath string
	commands []string
}

// NewHistory creates a new history manager
func NewHistory() *History {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		return nil
	}

	filePath := filepath.Join(homeDir, ".goterm_history")

	h := &History{
		filePath: filePath,
		commands: []string{},
	}

	// Load existing history
	h.load()

	return h
}

// load reads history from file
func (h *History) load() {
	if _, err := os.Stat(h.filePath); os.IsNotExist(err) {
		return
	}

	file, err := os.Open(h.filePath)
	if err != nil {
		fmt.Println("Error opening history file:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		cmd := scanner.Text()
		if cmd != "" {
			h.commands = append(h.commands, cmd)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading history file:", err)
	}
}

// Add adds a command to the history
func (h *History) Add(command string) {
	command = strings.TrimSpace(command)
	if command == "" {
		return
	}

	// Don't add duplicates consecutively
	if len(h.commands) > 0 && h.commands[len(h.commands)-1] == command {
		return
	}

	h.commands = append(h.commands, command)

	// Trim history if needed
	if len(h.commands) > HISTORY_LIMIT {
		h.commands = h.commands[len(h.commands)-HISTORY_LIMIT:]
	}

	// Save to file
	h.save()
}

// save writes history to file
func (h *History) save() {
	file, err := os.Create(h.filePath)
	if err != nil {
		fmt.Println("Error creating history file:", err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, cmd := range h.commands {
		fmt.Fprintln(writer, cmd)
	}

	writer.Flush()
}

// Show displays the command history
func (h *History) Show() {
	for i, cmd := range h.commands {
		fmt.Printf("%d: %s\n", i+1, cmd)
	}
}

// GetRecent returns the most recent command that starts with the given prefix
func (h *History) GetRecent(prefix string) string {
	if prefix == "" {
		return ""
	}

	for i := len(h.commands) - 1; i >= 0; i-- {
		if strings.HasPrefix(h.commands[i], prefix) {
			return h.commands[i]
		}
	}

	return ""
}

// GetHistoryPath returns the path to the history file
func (h *History) GetHistoryPath() string {
	return h.filePath
}

// GetAll returns all commands in the history
func (h *History) GetAll() []string {
	// Return a copy of the history slice to prevent modification of the original
	result := make([]string, len(h.commands))
	copy(result, h.commands)
	return result
}

func GetHistoryFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ".go-term_history"
	}
	return filepath.Join(homeDir, ".go-term_history")
}
