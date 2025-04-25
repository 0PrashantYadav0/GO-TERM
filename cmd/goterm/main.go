package main

import (
	"bufio"
	"fmt"
	"github/0PrashantYadav0/GO-TERM/internal/ai"
	"github/0PrashantYadav0/GO-TERM/internal/clipboard"
	"github/0PrashantYadav0/GO-TERM/internal/terminal"
	"github/0PrashantYadav0/GO-TERM/internal/ui"
	"github/0PrashantYadav0/GO-TERM/pkg/utils"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"github.com/peterh/liner"
)

func main() {
	// Clear console
	fmt.Print("\033[H\033[2J")

	// Print banner
	printBanner()

	// Start clipboard monitor in a goroutine
	suggestions := make(chan string)
	go clipboard.Monitor(suggestions)

	// Handle signals for clean exit
	setupSignalHandler()

	// Initialize history
	history := terminal.NewHistory()

	// Initialize liner for input with arrow key support
	line := liner.NewLiner()
	defer line.Close()

	// Configure liner
	line.SetCtrlCAborts(true)
	line.SetTabCompletionStyle(liner.TabPrints)

	// Load command history to liner
	if f, err := os.Open(terminal.GetHistoryFilePath()); err == nil {
		line.ReadHistory(f)
		f.Close()
	}

	spinner := ui.NewSpinner()

	for {
		// Display prompt - use a simple prompt that won't cause issues
		rawPrompt := ai.FormatPrompt()
		// Strip any ANSI color codes or other special characters
		safePrompt := utils.StripAnsi(rawPrompt)

		// Use a simple prompt if there are issues
		if strings.Contains(safePrompt, "\r") || strings.Contains(safePrompt, "\n") {
			safePrompt = "> "
		}

		// Check for clipboard suggestions (non-blocking)
		var suggestion string
		select {
		case suggestion = <-suggestions:
			// Got a suggestion
		default:
			// No suggestion available
		}

		// Read command with liner (supports arrow keys)
		var input string
		var err error

		// Use try-catch pattern for liner prompt
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("Recovered from liner panic, using simple prompt")
					fmt.Print(safePrompt)
					reader := bufio.NewReader(os.Stdin)
					inputBytes, _ := reader.ReadBytes('\n')
					input = strings.TrimSpace(string(inputBytes))
					err = nil
				}
			}()

			if suggestion != "" {
				// If we have a suggestion, show it separately to avoid prompt issues
				fmt.Print("\r" + safePrompt + suggestion)
				input, err = line.Prompt("")
				fmt.Print("\r\033[K") // Clear the line
			} else {
				input, err = line.Prompt(safePrompt)
			}
		}()

		if err == io.EOF {
			fmt.Println("\nExiting GO-TERM...")
			break
		} else if err != nil {
			if err.Error() == "Interrupted" {
				continue // Allow Ctrl+C to just cancel the current input
			}
			fmt.Printf("Simple mode activated due to error: %s\n", err)

			// Fall back to simple input method
			fmt.Print(safePrompt)
			reader := bufio.NewReader(os.Stdin)
			inputBytes, _ := reader.ReadBytes('\n')
			input = strings.TrimSpace(string(inputBytes))
		}

		if input == "" {
			continue
		}

		// Add to liner history
		line.AppendHistory(input)

		// Save history to file periodically
		if f, err := os.Create(terminal.GetHistoryFilePath()); err == nil {
			line.WriteHistory(f)
			f.Close()
		}

		// Handle exit command
		if input == "exit" {
			fmt.Println("Goodbye!")
			return
		}

		// Handle special commands
		if handled := handleSpecialCommands(input, history, spinner); handled {
			continue
		}

		// Add to our custom history
		history.Add(input)

		// Execute regular command
		terminal.ExecuteCommand(input)
	}
}

func printBanner() {
	cyan := color.New(color.FgCyan, color.Bold).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	width := utils.GetTerminalWidth()

	bannerText := []string{
		"╭───────────────────────────────────────────────╮",
		"│                                               │",
		"│   " + cyan("GO-TERM") + " - " + green("Intelligent Terminal Assistant") + "   │",
		"│                                               │",
		"│   " + yellow("✨ Powered by Gemini AI and Go ✨") + "         │",
		"│                                               │",
		"╰───────────────────────────────────────────────╯",
	}

	// Center the banner based on terminal width
	for _, line := range bannerText {
		padding := (width - len(utils.StripAnsi(line))) / 2
		if padding < 0 {
			padding = 0
		}
		fmt.Println(strings.Repeat(" ", padding) + line)
	}

	cmds := []string{
		"  • " + cyan("hm") + " - Get AI help for fixing the last error",
		"  • " + cyan("hp <query>") + " - Ask AI for a command",
		"  • " + cyan("he <query>") + " - Get AI explanation for a command",
	}

	fmt.Println()
	for _, cmd := range cmds {
		fmt.Println(cmd)
	}
	fmt.Println()
}

func handleSpecialCommands(input string, history *terminal.History, spinner *ui.Spinner) bool {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return false
	}

	cmd := parts[0]

	switch cmd {
	case "history":
		history.Show()
		return true

	case "cd":
		terminal.ChangeDirectory(input)
		return true

	case "cat":
		terminal.CatFile(input)
		return true

	case "hm": // Help Me (fix last error)
		spinner.Start("Processing last error...")
		result, err := ai.GenerateCommandForHm()
		spinner.Stop()

		if err != nil {
			fmt.Println("Error getting AI help:", err)
		} else if result == "3d8a19a704" {
			fmt.Println("Sorry, I couldn't help with that error.")
		} else {
			fmt.Println("Try:", result)
			// Offer to execute the command
			fmt.Print("Execute this command? [y/N]: ")
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(response)

			if strings.ToLower(response) == "y" {
				history.Add(result)
				terminal.ExecuteCommand(result)
			}
		}
		return true

	case "hp": // Help Please (get command suggestion)
		if len(parts) < 2 {
			fmt.Println("Usage: hp <your query>")
			return true
		}

		query := strings.Join(parts[1:], " ")
		spinner.Start("Processing your query...")
		result, err := ai.GenerateCommandForHp(query)
		spinner.Stop()

		if err != nil {
			fmt.Println("Error getting AI help:", err)
		} else if result == "3d8a19a704" {
			fmt.Println("Sorry, I couldn't generate a command for that query.")
		} else {
			fmt.Println("Try:", result)
			// Offer to execute the command
			fmt.Print("Execute this command? [y/N]: ")
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(response)

			if strings.ToLower(response) == "y" {
				history.Add(result)
				terminal.ExecuteCommand(result)
			}
		}
		return true

	case "he": // Help Explain
		if len(parts) < 2 {
			fmt.Println("Usage: he <your query>")
			return true
		}

		query := strings.Join(parts[1:], " ")
		spinner.Start("Getting explanation...")
		result, err := ai.ExplainCommand(query)
		spinner.Stop()

		if err != nil {
			fmt.Println("Error getting explanation:", err)
		} else if result == "3d8a19a704" {
			fmt.Println("Sorry, I couldn't provide an explanation.")
		} else {
			fmt.Println(result)
		}
		return true
	}

	return false
}

func setupSignalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\nExiting GO-TERM...")
		os.Exit(0)
	}()
}
