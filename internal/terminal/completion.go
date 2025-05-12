package terminal

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Completer handles command, flag, and path autocompletion
type Completer struct {
	commandCache map[string][]string // Command -> available flags
}

// NewCompleter initializes a new command completer
func NewCompleter() *Completer {
	return &Completer{
		commandCache: make(map[string][]string),
	}
}

// Complete attempts to complete the given input
func (c *Completer) Complete(input string) []string {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	// First word - command completion (including internal commands)
	if len(parts) == 1 {
		internalMatches := c.CompleteInternalCommands(parts[0])
		if len(internalMatches) > 0 {
			return internalMatches
		}
		return c.completeCommand(parts[0])
	}

	// Special completion for internal commands
	switch parts[0] {
	case "session", "sess":
		if len(parts) == 2 {
			subcommands := []string{"create", "switch", "list", "close", "layout"}
			return filterByPrefix(subcommands, parts[1])
		}
	case "alias", "a":
		if len(parts) == 2 {
			subcommands := []string{"add", "remove", "list"}
			return filterByPrefix(subcommands, parts[1])
		}
	case "bookmark", "bm":
		if len(parts) == 2 {
			subcommands := []string{"add", "remove", "list", "goto"}
			return filterByPrefix(subcommands, parts[1])
		}
	}

	// Continue with normal completion
	if strings.HasPrefix(parts[len(parts)-1], "-") {
		return c.completeFlags(parts[0], parts[len(parts)-1])
	}

	return c.completePath(parts[len(parts)-1])
}

// completeCommand completes command names
func (c *Completer) completeCommand(prefix string) []string {
	// Get commands from PATH
	// This is a simplified version - you'd want to cache this
	pathDirs := strings.Split(os.Getenv("PATH"), string(os.PathListSeparator))

	var matches []string
	for _, dir := range pathDirs {
		filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}

			if !d.IsDir() && strings.HasPrefix(d.Name(), prefix) {
				matches = append(matches, d.Name())
			}
			return nil
		})
	}

	return matches
}

// completeFlags completes command flags
func (c *Completer) completeFlags(cmd, prefix string) []string {
	// This would normally call the command with --help and parse the output
	// For now we'll return some dummy data for common commands

	if flags, ok := c.commandCache[cmd]; ok {
		var matches []string
		for _, flag := range flags {
			if strings.HasPrefix(flag, prefix) {
				matches = append(matches, flag)
			}
		}
		return matches
	}

	return nil
}

// completePath completes file paths
func (c *Completer) completePath(prefix string) []string {
	if !strings.Contains(prefix, "/") {
		prefix = "./" + prefix
	}

	dir := filepath.Dir(prefix)
	base := filepath.Base(prefix)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var matches []string
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, base) {
			if entry.IsDir() {
				matches = append(matches, filepath.Join(dir, name)+"/")
			} else {
				matches = append(matches, filepath.Join(dir, name))
			}
		}
	}

	return matches
}

// CompleteInternalCommands completes GO-TERM specific commands
func (c *Completer) CompleteInternalCommands(prefix string) []string {
	commands := []string{
		"hp",
		"he",
		"hm",
		"chat",
		"history",
		"exit",
		"session",
		"alias",
		"bookmark",
		"config",
		"update",
		"version",
		"help",
	}

	var matches []string
	for _, cmd := range commands {
		if strings.HasPrefix(cmd, prefix) {
			matches = append(matches, cmd)
		}
	}
	return matches
}

func filterByPrefix(items []string, prefix string) []string {
	var matches []string
	for _, item := range items {
		if strings.HasPrefix(item, prefix) {
			matches = append(matches, item)
		}
	}
	return matches
}
