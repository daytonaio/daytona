// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package pty

// Global PTY manager instance
var ptyManager = &PTYManager{
	sessions: make(map[string]*PTYSession),
}

// NewPTYManager creates a new PTY manager instance
func NewPTYManager() *PTYManager {
	return &PTYManager{
		sessions: make(map[string]*PTYSession),
	}
}

// Add adds a PTY session to the manager
func (m *PTYManager) Add(s *PTYSession) {
	m.mu.Lock()
	m.sessions[s.info.ID] = s
	m.mu.Unlock()
}

// Get retrieves a PTY session by ID
func (m *PTYManager) Get(id string) (*PTYSession, bool) {
	m.mu.RLock()
	s, ok := m.sessions[id]
	m.mu.RUnlock()
	return s, ok
}

// Delete removes a PTY session from the manager
func (m *PTYManager) Delete(id string) (*PTYSession, bool) {
	m.mu.Lock()
	s, ok := m.sessions[id]
	if ok {
		delete(m.sessions, id)
	}
	m.mu.Unlock()
	return s, ok
}

// List returns information about all managed PTY sessions
func (m *PTYManager) List() []PTYSessionInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]PTYSessionInfo, 0, len(m.sessions))
	for _, s := range m.sessions {
		out = append(out, s.Info())
	}
	return out
}
