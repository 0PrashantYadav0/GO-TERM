package terminal

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// Alias represents a command alias
type Alias struct {
	Name        string
	Command     string
	Description string
}

// AliasManager handles creation and usage of command aliases
type AliasManager struct {
	aliases     map[string]Alias
	configPath  string
	initialized bool
}

// NewAliasManager creates a new alias manager
func NewAliasManager(configDir string) *AliasManager {
	configPath := filepath.Join(configDir, "aliases.json")
	return &AliasManager{
		aliases:    make(map[string]Alias),
		configPath: configPath,
	}
}

// Initialize loads aliases from the config file
func (am *AliasManager) Initialize() error {
	if am.initialized {
		return nil
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(am.configPath)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return err
		}
	}

	// Load aliases if config file exists
	if _, err := os.Stat(am.configPath); !os.IsNotExist(err) {
		data, err := os.ReadFile(am.configPath)
		if err != nil {
			return err
		}

		var aliases []Alias
		if err := json.Unmarshal(data, &aliases); err != nil {
			return err
		}

		for _, alias := range aliases {
			am.aliases[alias.Name] = alias
		}
	}

	am.initialized = true
	return nil
}

// AddAlias creates a new command alias
func (am *AliasManager) AddAlias(name, command, description string) error {
	if err := am.Initialize(); err != nil {
		return err
	}

	if name == "" || command == "" {
		return errors.New("alias name and command cannot be empty")
	}

	am.aliases[name] = Alias{
		Name:        name,
		Command:     command,
		Description: description,
	}

	return am.saveAliases()
}

// RemoveAlias deletes an existing alias
func (am *AliasManager) RemoveAlias(name string) error {
	if err := am.Initialize(); err != nil {
		return err
	}

	if _, exists := am.aliases[name]; !exists {
		return errors.New("alias does not exist")
	}

	delete(am.aliases, name)
	return am.saveAliases()
}

// GetAlias retrieves an alias by name
func (am *AliasManager) GetAlias(name string) (Alias, error) {
	if err := am.Initialize(); err != nil {
		return Alias{}, err
	}

	alias, exists := am.aliases[name]
	if !exists {
		return Alias{}, errors.New("alias not found")
	}

	return alias, nil
}

// ListAliases returns all defined aliases
func (am *AliasManager) ListAliases() []Alias {
	if err := am.Initialize(); err != nil {
		return nil
	}

	aliases := make([]Alias, 0, len(am.aliases))
	for _, alias := range am.aliases {
		aliases = append(aliases, alias)
	}

	return aliases
}

// ExpandCommand expands any aliases in the given command
func (am *AliasManager) ExpandCommand(input string) string {
	if err := am.Initialize(); err != nil {
		return input
	}

	fields := strings.Fields(input)
	if len(fields) == 0 {
		return input
	}

	// Check if first word is an alias
	if alias, err := am.GetAlias(fields[0]); err == nil {
		// Replace the alias with its command
		expanded := alias.Command
		if len(fields) > 1 {
			// Append any arguments after the alias
			expanded += " " + strings.Join(fields[1:], " ")
		}
		return expanded
	}

	return input
}

// saveAliases saves all aliases to the config file
func (am *AliasManager) saveAliases() error {
	aliases := make([]Alias, 0, len(am.aliases))
	for _, alias := range am.aliases {
		aliases = append(aliases, alias)
	}

	data, err := json.MarshalIndent(aliases, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(am.configPath, data, 0644)
}
