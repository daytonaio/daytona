// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package interpreter

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ContextManager manages multiple interpreter contexts
type ContextManager struct {
	contexts   map[string]*InterpreterSession
	mu         sync.RWMutex
	defaultCwd string
}

var globalManager *ContextManager

// InitManager initializes the global context manager
func InitManager(defaultCwd string) {
	globalManager = &ContextManager{
		contexts:   make(map[string]*InterpreterSession),
		defaultCwd: defaultCwd,
	}
}

// CreateContext creates a new interpreter context
func (m *ContextManager) CreateContext(id, cwd, language string) (*InterpreterSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if context already exists
	if _, exists := m.contexts[id]; exists {
		return nil, fmt.Errorf("context with ID '%s' already exists", id)
	}

	// Validate language
	if language != "" && language != "python" {
		return nil, fmt.Errorf("unsupported language: %s (only 'python' is supported)", language)
	}
	if language == "" {
		language = "python"
	}

	// Use default cwd if not provided
	if cwd == "" {
		cwd = m.defaultCwd
	}

	// Create new session
	session := &InterpreterSession{
		info: InterpreterSessionInfo{
			ID:        id,
			Cwd:       cwd,
			CreatedAt: time.Now(),
			Active:    false,
			Language:  language,
		},
	}

	// Start the session
	if err := session.start(); err != nil {
		return nil, fmt.Errorf("failed to start context: %w", err)
	}

	// Store in map
	m.contexts[id] = session
	return session, nil
}

// GetContext retrieves an existing context by ID
func (m *ContextManager) GetContext(id string) (*InterpreterSession, error) {
	m.mu.RLock()
	session, exists := m.contexts[id]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("context with ID '%s' not found", id)
	}

	// Check if session is active, restart if needed
	info := session.Info()
	if !info.Active {
		// Session was stopped (e.g., exit() called), restart it
		if err := session.start(); err != nil {
			return nil, fmt.Errorf("failed to restart context '%s': %w", id, err)
		}
	}

	return session, nil
}

// GetOrCreateDefaultContext gets or creates the default context
func (m *ContextManager) GetOrCreateDefaultContext() (*InterpreterSession, error) {
	m.mu.RLock()
	session, exists := m.contexts["default"]
	m.mu.RUnlock()

	if exists {
		// Check if active, restart if needed
		info := session.Info()
		if !info.Active {
			if err := session.start(); err != nil {
				return nil, fmt.Errorf("failed to restart default context: %w", err)
			}
		}
		return session, nil
	}

	// Create default context
	return m.CreateContext("default", m.defaultCwd, "python")
}

// DeleteContext removes a context and shuts it down
func (m *ContextManager) DeleteContext(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.contexts[id]
	if !exists {
		return fmt.Errorf("context with ID '%s' not found", id)
	}

	// Shutdown the session
	session.shutdown()

	// Remove from map
	delete(m.contexts, id)
	return nil
}

// ListContexts returns information about all contexts
func (m *ContextManager) ListContexts() []InterpreterSessionInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	contexts := make([]InterpreterSessionInfo, 0, len(m.contexts))
	for _, session := range m.contexts {
		contexts = append(contexts, session.Info())
	}
	return contexts
}

// Global convenience functions

// CreateContext creates a new context using the global manager
func CreateContext(cwd, language string) (*InterpreterSession, error) {
	if globalManager == nil {
		return nil, fmt.Errorf("context manager not initialized")
	}
	// Generate unique ID
	id := uuid.NewString()
	return globalManager.CreateContext(id, cwd, language)
}

// GetContext gets a context by ID using the global manager
func GetContext(id string) (*InterpreterSession, error) {
	if globalManager == nil {
		return nil, fmt.Errorf("context manager not initialized")
	}
	return globalManager.GetContext(id)
}

// GetOrCreateDefaultContext gets or creates the default context
func GetOrCreateDefaultContext() (*InterpreterSession, error) {
	if globalManager == nil {
		return nil, fmt.Errorf("context manager not initialized")
	}
	return globalManager.GetOrCreateDefaultContext()
}

// DeleteContext deletes a context using the global manager
func DeleteContext(id string) error {
	if globalManager == nil {
		return fmt.Errorf("context manager not initialized")
	}
	return globalManager.DeleteContext(id)
}

// ListContexts lists all contexts using the global manager
func ListContexts() []InterpreterSessionInfo {
	if globalManager == nil {
		return []InterpreterSessionInfo{}
	}
	return globalManager.ListContexts()
}
