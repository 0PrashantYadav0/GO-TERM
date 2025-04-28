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
	"time"

	"github.com/fatih/color"
	"github.com/peterh/liner"
)

func main() {
	// Clear console
	fmt.Print("\033[H\033[2J")

	// Force enable colors
	ai.EnableColors()

	// Print enhanced banner with animation
	printEnhancedBanner()

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
	line.SetCtrlCAborts(false)
	line.SetTabCompletionStyle(liner.TabCircular)

	// Load command history to liner
	if f, err := os.Open(terminal.GetHistoryFilePath()); err == nil {
		line.ReadHistory(f)
		f.Close()
	}

	spinner := ui.NewSpinner()

	for {
		// Display colorful divider before each prompt
		printDivider()

		// Display prompt - use a simple prompt that won't cause issues
		rawPrompt := ai.FormatPrompt()
		// Strip any ANSI color codes or other special characters
		safePrompt := utils.StripAnsi(rawPrompt)

		// Use a simple prompt if there are issues
		if strings.Contains(safePrompt, "\r") || strings.Contains(safePrompt, "\n") {
			safePrompt = getColorfulPrompt()
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

			// Setup history suggestion completer
			line.SetCompleter(func(line string) (candidates []string) {
				if line == "" {
					return nil
				}

				// Find matching history command
				suggestion := history.GetRecent(line)
				if suggestion != "" && suggestion != line && strings.HasPrefix(suggestion, line) {
					return []string{suggestion}
				}
				return nil
			})

			if suggestion != "" {
				// If we have a clipboard suggestion, show it separately
				suggestedText := color.New(color.FgHiMagenta).Sprint(suggestion)
				fmt.Print("\r" + safePrompt + suggestedText)
				input, err = line.Prompt("")
				fmt.Print("\r\033[K") // Clear the line
			} else {
				// Simple prompt with tab completion from history
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
			errorText := color.New(color.FgHiRed).Sprintf("Simple mode activated due to error: %s", err)
			fmt.Println(errorText)

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
			printExitMessage()
			return
		}

		// Handle special commands
		if handled := handleSpecialCommands(input, history, spinner); handled {
			continue
		}

		// Add to our custom history
		history.Add(input)

		// Execute regular command with animated loading screen
		cmdDone := make(chan bool)

		// Start spinner in a goroutine
		spinner.Start(color.New(color.FgHiBlue, color.Bold).Sprint("⚡ Executing: ") +
			color.New(color.FgHiCyan).Sprint(input))

		// Execute the command in a goroutine
		go func() {
			terminal.ExecuteCommand(input)
			cmdDone <- true
		}()

		// Wait for command to complete
		<-cmdDone

		// Stop the spinner
		spinner.Stop()
	}
}

func printEnhancedBanner() {
	cyan := color.New(color.FgCyan, color.Bold).SprintFunc()
	green := color.New(color.FgGreen, color.Bold).SprintFunc()
	yellow := color.New(color.FgYellow, color.Bold).SprintFunc()
	magenta := color.New(color.FgMagenta, color.Bold).SprintFunc()
	blue := color.New(color.FgBlue, color.Bold).SprintFunc()

	width := utils.GetTerminalWidth()

	// Colorful top border with gradient effect
	topBorder := "╭" + strings.Repeat("━", width-2) + "╮"
	fmt.Println(blue(topBorder))

	// Logo with animation
	logoLines := []string{
		"",
		"   ██████╗  ██████╗     ████████╗███████╗██████╗ ███╗   ███╗",
		"  ██╔════╝ ██╔═══██╗    ╚══██╔══╝██╔════╝██╔══██╗████╗ ████║",
		"  ██║  ███╗██║   ██║       ██║   █████╗  ██████╔╝██╔████╔██║",
		"  ██║   ██║██║   ██║       ██║   ██╔══╝  ██╔══██╗██║╚██╔╝██║",
		"  ╚██████╔╝╚██████╔╝       ██║   ███████╗██║  ██║██║ ╚═╝ ██║",
		"   ╚═════╝  ╚═════╝        ╚═╝   ╚══════╝╚═╝  ╚═╝╚═╝     ╚═╝",
		"",
	}

	// Animate the logo
	for _, line := range logoLines {
		centerPadding := width/3
		if centerPadding < 0 {
			centerPadding = 0
		}
		fmt.Print(strings.Repeat(" ", centerPadding))
		// Animate each character
		for _, char := range line {
			switch char {
			case '█':
				fmt.Print(cyan(string(char)))
			case '╔', '╗', '╚', '╝', '═', '║':
				fmt.Print(magenta(string(char)))
			default:
				fmt.Print(string(char))
			}
			time.Sleep(1 * time.Millisecond)
		}
		fmt.Println()
	}

	// Tagline with sparkle
	tagline := "✨ " + green("Intelligent Terminal Assistant") + " - " + yellow("Powered by Gemini AI") + " ✨"
	centerPadding := (width - len(utils.StripAnsi(tagline))) / 2
	if centerPadding < 0 {
		centerPadding = 0
	}
	fmt.Println(strings.Repeat(" ", centerPadding) + tagline)

	// Version info
	versionInfo := magenta("v1.0.0") + " - " + yellow("Your AI-powered command line companion")
	centerPadding = (width - len(utils.StripAnsi(versionInfo))) / 2
	if centerPadding < 0 {
		centerPadding = 0
	}
	fmt.Println(strings.Repeat(" ", centerPadding) + versionInfo)
	fmt.Println()

	// Bottom border with gradient
	bottomBorder := "╰" + strings.Repeat("━", width-2) + "╯"
	fmt.Println(blue(bottomBorder))

	// Command help section with improved formatting
	fmt.Println()
	helpTitle := "📋 " + yellow("Available Commands:") + " 📋"
	centerPadding = (width - len(utils.StripAnsi(helpTitle))) / 2
	if centerPadding < 0 {
		centerPadding = 0
	}
	fmt.Println(strings.Repeat(" ", centerPadding) + helpTitle)

	cmds := []string{
		"  • " + cyan("hm") + " - " + green("Get AI help for fixing the last error"),
		"  • " + cyan("hp <query>") + " - " + green("Ask AI for a command"),
		"  • " + cyan("he <query>") + " - " + green("Get AI explanation for a command"),
		"  • " + cyan("history") + " - " + green("Show command history"),
		"  • " + cyan("exit") + " - " + green("Exit GO-TERM"),
	}

	for _, cmd := range cmds {
		fmt.Println(cmd)
		time.Sleep(50 * time.Millisecond) // Small delay for animation effect
	}

	fmt.Println()
}

func printDivider() {
	width := utils.GetTerminalWidth()
	blue := color.New(color.FgBlue, color.Bold).SprintFunc()
	bottomBorder := "╰" + strings.Repeat("━", width-2) + "╯"
	fmt.Println(blue(bottomBorder))
}

func getColorfulPrompt() string {
	// Get current directory for the prompt
	dir, _ := os.Getwd()
	if home, err := os.UserHomeDir(); err == nil {
		dir = strings.Replace(dir, home, "~", 1)
	}

	// Create colorful prompt components
	timestamp := color.New(color.FgHiBlack).Sprintf("[%s]", time.Now().Format("15:04:05"))
	directory := color.New(color.FgHiBlue, color.Bold).Sprintf("%s", dir)
	promptChar := color.New(color.FgHiMagenta, color.Bold).Sprint(" ❯ ")

	return timestamp + " " + directory + promptChar
}

func printExitMessage() {
	// Create a colorful exit message with animation
	exitMsg := "Thank you for using GO-TERM! Goodbye!"
	fmt.Println()

	// Print characters with rainbow colors one by one
	colorFuncs := []func(a ...interface{}) string{
		color.New(color.FgRed, color.Bold).SprintFunc(),
		color.New(color.FgYellow, color.Bold).SprintFunc(),
		color.New(color.FgGreen, color.Bold).SprintFunc(),
		color.New(color.FgCyan, color.Bold).SprintFunc(),
		color.New(color.FgBlue, color.Bold).SprintFunc(),
		color.New(color.FgMagenta, color.Bold).SprintFunc(),
	}

	for i, char := range exitMsg {
		colorIndex := i % len(colorFuncs)
		fmt.Print(colorFuncs[colorIndex](string(char)))
		time.Sleep(30 * time.Millisecond)
	}
	fmt.Println("\n")
}

func handleSpecialCommands(input string, history *terminal.History, spinner *ui.Spinner) bool {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return false
	}

	cmd := parts[0]
	successColor := color.New(color.FgGreen, color.Bold).SprintFunc()
	errorColor := color.New(color.FgRed, color.Bold).SprintFunc()
	headerColor := color.New(color.FgMagenta, color.Bold).SprintFunc()

	switch cmd {
	case "history":
		fmt.Println(headerColor("📜 Command History:"))
		history.Show()
		return true

	case "cd":
		terminal.ChangeDirectory(input)
		return true

	case "cat":
		terminal.CatFile(input)
		return true

	case "hm": // Help Me (fix last error)
		spinner.Start(color.New(color.FgCyan).Sprint("✨ Processing last error..."))
		result, err := ai.GenerateCommandForHm()
		spinner.Stop()

		if err != nil {
			fmt.Println(errorColor("Error getting AI help:"), err)
		} else if result == "3d8a19a704" {
			fmt.Println(errorColor("Sorry, I couldn't help with that error."))
		} else {
			fmt.Println(headerColor("🚀 Try:"), color.New(color.FgHiCyan, color.Bold).Sprint(result))

			// Copy the command to clipboard
			if err := clipboard.Write(result); err == nil {
				fmt.Println(successColor("✓ Command copied to clipboard"))
			}
		}
		return true

	case "hp": // Help Please (get command suggestion)
		if len(parts) < 2 {
			fmt.Println(errorColor("Usage:"), "hp <your query>")
			return true
		}

		query := strings.Join(parts[1:], " ")
		spinner.Start(color.New(color.FgCyan).Sprint("✨ Processing your query..."))
		result, err := ai.GenerateCommandForHp(query)
		spinner.Stop()

		if err != nil {
			fmt.Println(errorColor("Error getting AI help:"), err)
		} else if result == "3d8a19a704" {
			fmt.Println(errorColor("Sorry, I couldn't generate a command for that query."))
		} else {
			fmt.Println(headerColor("🚀 Try:"), color.New(color.FgHiCyan, color.Bold).Sprint(result))

			// Copy the command to clipboard
			if err := clipboard.Write(result); err == nil {
				fmt.Println(successColor("✓ Command copied to clipboard"))
			}
		}
		return true

	case "he": // Help Explain
		if len(parts) < 2 {
			fmt.Println(errorColor("Usage:"), "he <your query>")
			return true
		}

		query := strings.Join(parts[1:], " ")
		spinner.Start(color.New(color.FgCyan).Sprint("✨ Getting explanation..."))
		result, err := ai.ExplainCommand(query)
		spinner.Stop()

		fmt.Println(headerColor("📚 Explanation:"))

		if err != nil {
			fmt.Println(errorColor("Error getting explanation:"), err)
		} else if result == "3d8a19a704" {
			fmt.Println(errorColor("Sorry, I couldn't provide an explanation."))
		} else {
			// Print the explanation in a box
			width := utils.GetTerminalWidth()
			boxWidth := width - 4

			// Top border
			fmt.Println(color.New(color.FgHiBlack).Sprint("┌" + strings.Repeat("─", boxWidth) + "┐"))

			// Split explanation into lines and print with padding
			explLines := strings.Split(result, "\n")
			for _, line := range explLines {
				// Handle line wrapping for long lines
				for len(line) > boxWidth-4 {
					fmt.Print(color.New(color.FgHiBlack).Sprint("│ "))
					fmt.Print(color.New(color.FgHiWhite).Sprint(line[:boxWidth-4]))
					fmt.Println(color.New(color.FgHiBlack).Sprint(" │"))
					line = line[boxWidth-4:]
				}
				fmt.Print(color.New(color.FgHiBlack).Sprint("│ "))
				fmt.Print(color.New(color.FgHiWhite).Sprint(line))
				padding := boxWidth - 2 - len(line)
				fmt.Print(strings.Repeat(" ", padding))
				fmt.Println(color.New(color.FgHiBlack).Sprint(" │"))
			}

			// Bottom border
			fmt.Println(color.New(color.FgHiBlack).Sprint("└" + strings.Repeat("─", boxWidth) + "┘"))

			// For explanations, we might want to copy them as well
			if err := clipboard.Write(result); err == nil {
				fmt.Println(successColor("✓ Explanation copied to clipboard"))
			}
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
		fmt.Println("\n" + color.New(color.FgYellow, color.Bold).Sprint("Exiting GO-TERM..."))
		os.Exit(0)
	}()
}
