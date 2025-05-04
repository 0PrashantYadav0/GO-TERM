package terminal

import (
	"errors"
	"os/exec"
	"sync"
)

// Session represents a terminal session
type Session struct {
	ID      int
	Name    string
	Command *exec.Cmd
	Active  bool
}

// Layout types for arranging terminal sessions
type LayoutType int

const (
	Tabs LayoutType = iota
	VerticalSplit
	HorizontalSplit
	Grid
)

// Multiplexer manages multiple terminal sessions
type Multiplexer struct {
	sessions     map[int]*Session
	nextID       int
	activeID     int
	layout       LayoutType
	sessionMutex sync.Mutex
}

// NewMultiplexer creates a new terminal multiplexer
func NewMultiplexer() *Multiplexer {
	return &Multiplexer{
		sessions: make(map[int]*Session),
		nextID:   1,
		layout:   Tabs,
	}
}

// CreateSession creates a new terminal session
func (m *Multiplexer) CreateSession(name string, command string, args ...string) (int, error) {
	m.sessionMutex.Lock()
	defer m.sessionMutex.Unlock()

	cmd := exec.Command(command, args...)

	session := &Session{
		ID:      m.nextID,
		Name:    name,
		Command: cmd,
		Active:  false,
	}

	m.sessions[m.nextID] = session
	m.nextID++

	// If this is the first session, make it active
	if len(m.sessions) == 1 {
		m.activeID = session.ID
		session.Active = true
	}

	return session.ID, nil
}

// SwitchToSession changes the active session
func (m *Multiplexer) SwitchToSession(id int) error {
	m.sessionMutex.Lock()
	defer m.sessionMutex.Unlock()

	if _, exists := m.sessions[id]; !exists {
		return errors.New("session not found")
	}

	// Deactivate the current session
	if active, ok := m.sessions[m.activeID]; ok {
		active.Active = false
	}

	// Activate the requested session
	m.sessions[id].Active = true
	m.activeID = id

	return nil
}

// RemoveSession terminates and removes a session
func (m *Multiplexer) RemoveSession(id int) error {
	m.sessionMutex.Lock()
	defer m.sessionMutex.Unlock()

	session, exists := m.sessions[id]
	if !exists {
		return errors.New("session not found")
	}

	// Terminate the process if it's running
	if session.Command.Process != nil {
		_ = session.Command.Process.Kill()
	}

	// Remove from sessions map
	delete(m.sessions, id)

	// If we removed the active session, activate another one if available
	if id == m.activeID && len(m.sessions) > 0 {
		for newID := range m.sessions {
			m.sessions[newID].Active = true
			m.activeID = newID
			break
		}
	}

	return nil
}

// ListSessions returns all current sessions
func (m *Multiplexer) ListSessions() []*Session {
	m.sessionMutex.Lock()
	defer m.sessionMutex.Unlock()

	sessions := make([]*Session, 0, len(m.sessions))
	for _, session := range m.sessions {
		sessions = append(sessions, session)
	}

	return sessions
}

// SetLayout changes how sessions are displayed
func (m *Multiplexer) SetLayout(layout LayoutType) {
	m.layout = layout
}

// GetLayout returns the current layout
func (m *Multiplexer) GetLayout() LayoutType {
	return m.layout
}

// GetActiveSession returns the currently active session
func (m *Multiplexer) GetActiveSession() (*Session, error) {
	m.sessionMutex.Lock()
	defer m.sessionMutex.Unlock()

	if session, exists := m.sessions[m.activeID]; exists {
		return session, nil
	}

	return nil, errors.New("no active session")
}
