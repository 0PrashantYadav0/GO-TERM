package terminal

import (
	"encoding/json"
	"fmt"
	"github/0PrashantYadav0/GO-TERM/pkg/logger"
	"github/0PrashantYadav0/GO-TERM/pkg/utils"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"
	"time"
)

type CommandLog struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	Command   struct {
		Raw        string   `json:"raw"`
		Executable string   `json:"executable"`
		Arguments  []string `json:"arguments"`
		CWD        string   `json:"cwd"`
	} `json:"command"`
	Output struct {
		Stderr   string `json:"stderr"`
		ExitCode int    `json:"exitCode"`
		Error    string `json:"error,omitempty"`
	} `json:"output"`
	Metadata struct {
		User     string `json:"user"`
		Platform string `json:"platform"`
		Shell    string `json:"shell"`
	} `json:"metadata"`
}

// ExecuteCommand executes a shell command and handles the result
func ExecuteCommand(input string) {
	logEntry := initCommandLog(input)

	parts := strings.Fields(input)
	if len(parts) == 0 {
		return
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	// Capture stderr separately
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println("Error setting up stderr pipe:", err)
		return
	}

	err = cmd.Start()
	if err != nil {
		logEntry.Output.Error = err.Error()
		logEntry.Output.ExitCode = 1
		fmt.Println("Error starting command:", err)
		saveCommandLog(logEntry)
		return
	}

	// Read stderr
	stderrBytes, _ := io.ReadAll(stderrPipe)
	if len(stderrBytes) > 0 {
		logEntry.Output.Stderr = string(stderrBytes)
	}

	err = cmd.Wait()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			logEntry.Output.ExitCode = exitErr.ExitCode()
		} else {
			logEntry.Output.ExitCode = 1
			logEntry.Output.Error = err.Error()
		}
	}

	// Save to error log file if there was an error
	if logEntry.Output.ExitCode != 0 || logEntry.Output.Stderr != "" {
		saveCommandLog(logEntry)
	}
}

// SafeExecuteCommand executes a command with better error recovery
func SafeExecuteCommand(input string) {
	defer func() {
		if r := recover(); r != nil {
			stackTrace := string(debug.Stack())
			logger.Error("Command execution panic: %v\n%s", r, stackTrace)

			// Show user-friendly message
			fmt.Printf("%s Command execution failed unexpectedly.\n", utils.GetErrorMessageFace())
			fmt.Println("The error has been logged. Please report this issue.")
		}
	}()

	ExecuteCommand(input)
}

// EnhancedExecuteCommand executes commands with retries and comprehensive error handling
func EnhancedExecuteCommand(input string, retries int) {

	for attempt := 0; attempt <= retries; attempt++ {
		logEntry := initCommandLog(input)

		parts := strings.Fields(input)
		if len(parts) == 0 {
			return
		}

		// Check if command exists before trying to execute it
		if _, err := exec.LookPath(parts[0]); err != nil {
			fmt.Printf("Command not found: %s\n", parts[0])

			// Suggest alternatives
			suggestions := getSimilarCommands(parts[0])
			if len(suggestions) > 0 {
				fmt.Println("Did you mean one of these?")
				for _, suggestion := range suggestions {
					fmt.Printf("  %s\n", suggestion)
				}
			}
			return
		}

		// Execute as in the original function...
		cmd := exec.Command(parts[0], parts[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout

		// Capture stderr separately
		stderrPipe, err := cmd.StderrPipe()
		if err != nil {
			logger.Error("Error setting up stderr pipe: %v", err)
			fmt.Println("Error setting up command:", err)
			continue
		}

		err = cmd.Start()
		if err != nil {
			logEntry.Output.Error = err.Error()
			logEntry.Output.ExitCode = 1
			logger.Error("Error starting command: %v", err)
			fmt.Println("Error starting command:", err)

			if attempt < retries {
				fmt.Printf("Retrying... (%d/%d)\n", attempt+1, retries)
				time.Sleep(time.Second)
				continue
			}
			saveCommandLog(logEntry)
			return
		}

		// Read stderr
		stderrBytes, _ := io.ReadAll(stderrPipe)
		if len(stderrBytes) > 0 {
			logEntry.Output.Stderr = string(stderrBytes)
		}

		err = cmd.Wait()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				logEntry.Output.ExitCode = exitErr.ExitCode()
			} else {
				logEntry.Output.ExitCode = 1
				logEntry.Output.Error = err.Error()
			}
		}

		// Save to error log file if there was an error
		if logEntry.Output.ExitCode != 0 || logEntry.Output.Stderr != "" {
			saveCommandLog(logEntry)
		}

		// If command executed successfully or it's a legitimate failure (not connection/IO error)
		if err == nil || logEntry.Output.ExitCode != 0 {
			return
		}

		if attempt < retries {
			fmt.Printf("Command failed, retrying... (%d/%d)\n", attempt+1, retries)
			time.Sleep(time.Second)
		}
	}
}

// Helper to find similar commands
func getSimilarCommands(cmd string) []string {
	commonCommands := []string{
		"ls", "cd", "pwd", "cp", "mv", "rm", "mkdir", "rmdir",
		"cat", "grep", "find", "echo", "touch", "chmod", "chown",
		"git", "docker", "npm", "go", "python", "pip", "brew",
	}

	var similar []string
	for _, command := range commonCommands {
		if levenshteinDistance(cmd, command) <= 2 {
			similar = append(similar, command)
		}
	}
	return similar
}

// levenshteinDistance calculates the edit distance between two strings
func levenshteinDistance(s1, s2 string) int {
	// Implementation of the Levenshtein distance algorithm
	// (simplified version shown here)
	if s1 == s2 {
		return 0
	}

	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	if s1[0] == s2[0] {
		return levenshteinDistance(s1[1:], s2[1:])
	}

	return 1 + min(
		levenshteinDistance(s1[1:], s2), // deletion
		min(
			levenshteinDistance(s1, s2[1:]),     // insertion
			levenshteinDistance(s1[1:], s2[1:]), // substitution
		),
	)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func initCommandLog(command string) *CommandLog {
	parts := strings.Fields(command)

	log := &CommandLog{
		ID:        generateID(),
		Timestamp: time.Now().Format(time.RFC3339),
	}

	log.Command.Raw = command
	if len(parts) > 0 {
		log.Command.Executable = parts[0]
		log.Command.Arguments = parts[1:]
	}
	log.Command.CWD = getCurrentDir()

	log.Output.ExitCode = 0

	log.Metadata.User = utils.GetUsername()
	log.Metadata.Platform = utils.GetPlatform()
	log.Metadata.Shell = utils.GetShellName()

	return log
}

func generateID() string {
	return fmt.Sprintf("cmd_%d_%s", time.Now().Unix(), utils.RandomString(8))
}

func saveCommandLog(log *CommandLog) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		return
	}

	filePath := filepath.Join(homeDir, ".goterm_error")

	// Read existing logs
	var logs []CommandLog
	if fileData, err := os.ReadFile(filePath); err == nil {
		_ = json.Unmarshal(fileData, &logs)
	}

	// Add new log
	logs = append(logs, *log)

	// Keep only the last 10 logs
	if len(logs) > 10 {
		logs = logs[len(logs)-10:]
	}

	// Write back to file
	fileData, err := json.MarshalIndent(logs, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling logs:", err)
		return
	}

	err = os.WriteFile(filePath, fileData, 0644)
	if err != nil {
		fmt.Println("Error saving logs:", err)
	}
}

func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return dir
}

// ChangeDirectory changes the current working directory
func ChangeDirectory(input string) {
	parts := strings.Fields(input)

	var dir string
	if len(parts) == 1 {
		// cd with no arguments goes to home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Error getting home directory:", err)
			return
		}
		dir = homeDir
	} else {
		dir = parts[1]
	}

	err := os.Chdir(dir)
	if err != nil {
		fmt.Println("Error changing directory:", err)
	}
}

// StripAnsi removes ANSI escape codes from a string
func StripAnsi(str string) string {
	ansi := regexp.MustCompile("\033\\[(?:[0-9]{1,3}(?:;[0-9]{1,3})*)?[m|K]")
	return ansi.ReplaceAllString(str, "")
}

// CatFile displays the contents of a file
func CatFile(input string) {
	parts := strings.Fields(input)
	if len(parts) < 2 {
		fmt.Println("Usage: cat <filename>")
		return
	}

	filename := parts[1]
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	fmt.Println(string(data))
}
