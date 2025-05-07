package terminal

import (
	"fmt"
	"os"
	"strings"
)

// CLI handles command line interactions for the terminal features
type CLI struct {
	mux       *Multiplexer
	aliases   *AliasManager
	bookmarks *BookmarkManager
}

// NewCLI creates a new CLI handler
func NewCLI(mux *Multiplexer, aliases *AliasManager, bookmarks *BookmarkManager) *CLI {
	return &CLI{
		mux:       mux,
		aliases:   aliases,
		bookmarks: bookmarks,
	}
}

// HandleCommand processes internal commands
func (c *CLI) HandleCommand(cmd string, args []string) (bool, error) {
	switch cmd {
	case "session", "sess":
		return true, c.handleSessionCommand(args)
	case "alias", "a":
		return true, c.handleAliasCommand(args)
	case "bookmark", "bm":
		return true, c.handleBookmarkCommand(args)
	default:
		return false, nil
	}
}

// handleSessionCommand manages terminal sessions
func (c *CLI) handleSessionCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("session command requires a subcommand")
	}

	switch args[0] {
	case "create", "new":
		if len(args) < 2 {
			return fmt.Errorf("session create requires a name")
		}
		name := args[1]
		command := os.Getenv("SHELL")
		if len(args) >= 3 {
			command = args[2]
		}

		sessionID, err := c.mux.CreateSession(name, command)
		if err != nil {
			return err
		}
		fmt.Printf("Created session %d (%s)\n", sessionID, name)

	case "switch", "sw":
		if len(args) < 2 {
			return fmt.Errorf("session switch requires a session ID")
		}
		var id int
		if _, err := fmt.Sscanf(args[1], "%d", &id); err != nil {
			return err
		}

		if err := c.mux.SwitchToSession(id); err != nil {
			return err
		}
		fmt.Printf("Switched to session %d\n", id)

	case "list", "ls":
		sessions := c.mux.ListSessions()
		fmt.Println("Available sessions:")
		for _, session := range sessions {
			activeStr := ""
			if session.Active {
				activeStr = " (active)"
			}
			fmt.Printf("  %d: %s%s\n", session.ID, session.Name, activeStr)
		}

	case "close", "rm":
		if len(args) < 2 {
			return fmt.Errorf("session close requires a session ID")
		}
		var id int
		if _, err := fmt.Sscanf(args[1], "%d", &id); err != nil {
			return err
		}

		if err := c.mux.RemoveSession(id); err != nil {
			return err
		}
		fmt.Printf("Closed session %d\n", id)

	case "layout":
		if len(args) < 2 {
			return fmt.Errorf("session layout requires a layout type")
		}

		var layout LayoutType
		switch strings.ToLower(args[1]) {
		case "tabs":
			layout = Tabs
		case "vsplit":
			layout = VerticalSplit
		case "hsplit":
			layout = HorizontalSplit
		case "grid":
			layout = Grid
		default:
			return fmt.Errorf("unknown layout: %s", args[1])
		}

		c.mux.SetLayout(layout)
		fmt.Printf("Changed layout to %s\n", args[1])

	default:
		return fmt.Errorf("unknown session subcommand: %s", args[0])
	}

	return nil
}

// handleAliasCommand manages command aliases
func (c *CLI) handleAliasCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("alias command requires a subcommand")
	}

	switch args[0] {
	case "add", "new":
		if len(args) < 3 {
			return fmt.Errorf("alias add requires a name and command")
		}
		name := args[1]
		command := args[2]
		description := ""
		if len(args) >= 4 {
			description = args[3]
		}

		if err := c.aliases.AddAlias(name, command, description); err != nil {
			return err
		}
		fmt.Printf("Added alias %s -> %s\n", name, command)

	case "remove", "rm":
		if len(args) < 2 {
			return fmt.Errorf("alias remove requires a name")
		}
		name := args[1]

		if err := c.aliases.RemoveAlias(name); err != nil {
			return err
		}
		fmt.Printf("Removed alias %s\n", name)

	case "list", "ls":
		aliases := c.aliases.ListAliases()
		fmt.Println("Defined aliases:")
		for _, alias := range aliases {
			if alias.Description != "" {
				fmt.Printf("  %s -> %s (%s)\n", alias.Name, alias.Command, alias.Description)
			} else {
				fmt.Printf("  %s -> %s\n", alias.Name, alias.Command)
			}
		}

	default:
		return fmt.Errorf("unknown alias subcommand: %s", args[0])
	}

	return nil
}

// handleBookmarkCommand manages directory bookmarks
func (c *CLI) handleBookmarkCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("bookmark command requires a subcommand")
	}

	switch args[0] {
	case "add", "new":
		if len(args) < 3 {
			return fmt.Errorf("bookmark add requires a name and path")
		}
		name := args[1]
		path := args[2]
		description := ""
		if len(args) >= 4 {
			description = args[3]
		}

		if err := c.bookmarks.AddBookmark(name, path, description); err != nil {
			return err
		}
		fmt.Printf("Added bookmark %s -> %s\n", name, path)

	case "remove", "rm":
		if len(args) < 2 {
			return fmt.Errorf("bookmark remove requires a name")
		}
		name := args[1]

		if err := c.bookmarks.RemoveBookmark(name); err != nil {
			return err
		}
		fmt.Printf("Removed bookmark %s\n", name)

	case "list", "ls":
		bookmarks := c.bookmarks.ListBookmarks()
		fmt.Println("Defined bookmarks:")
		for _, bookmark := range bookmarks {
			if bookmark.Description != "" {
				fmt.Printf("  %s -> %s (%s)\n", bookmark.Name, bookmark.Path, bookmark.Description)
			} else {
				fmt.Printf("  %s -> %s\n", bookmark.Name, bookmark.Path)
			}
		}

	case "goto", "cd":
		if len(args) < 2 {
			return fmt.Errorf("bookmark goto requires a bookmark name")
		}
		name := args[1]

		bookmark, err := c.bookmarks.GetBookmark(name)
		if err != nil {
			return err
		}

		if err := os.Chdir(bookmark.Path); err != nil {
			return err
		}
		fmt.Printf("Changed directory to %s\n", bookmark.Path)

	default:
		return fmt.Errorf("unknown bookmark subcommand: %s", args[0])
	}

	return nil
}
