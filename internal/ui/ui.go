package ui

import (
    "fmt"
    "os"
    "time"
    
    "github.com/fatih/color"
)

// Spinner is a terminal spinner for displaying loading progress
type Spinner struct {
    frames      []frame
    interval    time.Duration
    isSpinning  bool
    stopChan    chan struct{}
    stream      *os.File
    frameIndex  int
}

type frame struct {
    animation string
    state     string
    color     func(...interface{}) string
}

// NewSpinner creates a new spinner
func NewSpinner() *Spinner {
    cyan := color.New(color.FgCyan).SprintFunc()
    blue := color.New(color.FgBlue).SprintFunc()
    
    frames := []frame{
        // Brain processing animation
        {animation: "⟨ ◠◡◠ ⟩", state: "  think", color: cyan},
        {animation: "⟨ ◡◠◡ ⟩", state: "  learn", color: cyan},
        {animation: "⟨ ◠◡◠ ⟩", state: " neural", color: cyan},
        {animation: "⟨ ◡◠◡ ⟩", state: "  sync ", color: cyan},

        // Circuit flow animation
        {animation: "[←↔→]", state: " parse", color: cyan},
        {animation: "[→↔←]", state: "  map ", color: cyan},
        {animation: "[←↔→]", state: " flow ", color: cyan},
        {animation: "[→↔←]", state: " link ", color: cyan},

        // Data pulse animation
        {animation: "(∙∵∙)", state: " data ", color: blue},
        {animation: "(∙∴∙)", state: " proc ", color: blue},
        {animation: "(∙∵∙)", state: " calc ", color: blue},
        {animation: "(∙∴∙)", state: " eval ", color: blue},

        // Matrix scan animation
        {animation: "⟦░▒▓⟧", state: " scan ", color: blue},
        {animation: "⟦▒▓░⟧", state: " read ", color: blue},
        {animation: "⟦▓░▒⟧", state: " load ", color: blue},
        {animation: "⟦░▒▓⟧", state: " feed ", color: blue},
    }
    
    return &Spinner{
        frames:     frames,
        interval:   100 * time.Millisecond,
        isSpinning: false,
        stopChan:   make(chan struct{}),
        stream:     os.Stdout,
        frameIndex: 0,
    }
}

// Start begins the spinner animation
func (s *Spinner) Start(text string) {
    if s.isSpinning {
        return
    }
    
    s.isSpinning = true
    s.frameIndex = 0
    
    // Hide cursor
    fmt.Fprint(s.stream, "\033[?25l")
    
    go func() {
        for {
            select {
            case <-s.stopChan:
                return
            default:
                s.render(text)
                time.Sleep(s.interval)
                s.frameIndex = (s.frameIndex + 1) % len(s.frames)
            }
        }
    }()
}

func (s *Spinner) render(text string) {
    frame := s.frames[s.frameIndex]
    
    // Clear line and move cursor to beginning
    fmt.Fprintf(s.stream, "\r\033[K")
    
    // Write frame with color
    fmt.Fprintf(s.stream, "%s%s %s", 
        frame.color(frame.animation), 
        frame.color(frame.state), 
        text)
}

// Stop stops the spinner animation
func (s *Spinner) Stop() {
    if !s.isSpinning {
        return
    }
    
    s.isSpinning = false
    s.stopChan <- struct{}{}
    
    // Clear line and move cursor to beginning
    fmt.Fprintf(s.stream, "\r\033[K")
    
    // Show cursor
    fmt.Fprint(s.stream, "\033[?25h")
}