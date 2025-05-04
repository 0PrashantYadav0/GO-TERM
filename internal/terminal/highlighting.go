package terminal

import (
	"regexp"
	"strings"
)

// Color ANSI escape codes
const (
	Reset   = "\033[0m"
	Bold    = "\033[1m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
)

// TokenType represents different parts of a command
type TokenType int

const (
	Command TokenType = iota
	Flag
	Argument
	Pipe
	Redirection
	Variable
	QuotedString
	Comment
)

// Token represents a part of a command with its type
type Token struct {
	Type  TokenType
	Value string
}

// Highlighter handles syntax highlighting for terminal commands
type Highlighter struct {
	patterns map[TokenType]*regexp.Regexp
}

// NewHighlighter creates a new syntax highlighter
func NewHighlighter() *Highlighter {
	h := &Highlighter{
		patterns: make(map[TokenType]*regexp.Regexp),
	}

	// Initialize regex patterns for different token types
	h.patterns[Command] = regexp.MustCompile(`^[a-zA-Z0-9_\-\.]+`)
	h.patterns[Flag] = regexp.MustCompile(`^-{1,2}[a-zA-Z0-9_\-]+`)
	h.patterns[Pipe] = regexp.MustCompile(`^\|`)
	h.patterns[Redirection] = regexp.MustCompile(`^[><]{1,2}`)
	h.patterns[Variable] = regexp.MustCompile(`^\$[a-zA-Z0-9_]+`)
	h.patterns[QuotedString] = regexp.MustCompile(`^(['"])(.*?)(\1)`)
	h.patterns[Comment] = regexp.MustCompile(`^#.*$`)

	return h
}

// Tokenize splits a command into tokens
func (h *Highlighter) Tokenize(command string) []Token {
	var tokens []Token
	remaining := strings.TrimSpace(command)

	for len(remaining) > 0 {
		matched := false

		// Skip whitespace
		if spaces := strings.TrimLeft(remaining, " \t"); len(spaces) != len(remaining) {
			remaining = spaces
			continue
		}

		// Try to match each token type
		for tokenType, pattern := range h.patterns {
			if match := pattern.FindString(remaining); match != "" {
				tokens = append(tokens, Token{Type: tokenType, Value: match})
				remaining = strings.TrimPrefix(remaining, match)
				matched = true
				break
			}
		}

		// If no match, treat as argument
		if !matched {
			endIdx := strings.IndexAny(remaining, " \t|><")
			if endIdx == -1 {
				endIdx = len(remaining)
			}
			tokens = append(tokens, Token{Type: Argument, Value: remaining[:endIdx]})
			remaining = remaining[endIdx:]
		}
	}

	return tokens
}

// Highlight adds color to a command string
func (h *Highlighter) Highlight(command string) string {
	tokens := h.Tokenize(command)
	result := ""

	for _, token := range tokens {
		colored := h.colorizeToken(token)
		result += colored
	}

	return result + Reset
}

// colorizeToken applies colors based on token type
func (h *Highlighter) colorizeToken(token Token) string {
	switch token.Type {
	case Command:
		return Bold + Green + token.Value + Reset
	case Flag:
		return Yellow + token.Value + Reset
	case Pipe, Redirection:
		return Bold + Magenta + token.Value + Reset
	case Variable:
		return Cyan + token.Value + Reset
	case QuotedString:
		return Red + token.Value + Reset
	case Comment:
		return Blue + token.Value + Reset
	default: // Argument
		return White + token.Value + Reset
	}
}
