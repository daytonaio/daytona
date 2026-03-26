// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package interpreter

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/google/uuid"
)

// Manager manages multiple interpreter contexts
type Manager struct {
	contexts   map[string]*Context
	mu         sync.RWMutex
	defaultCwd string
}

var globalManager *Manager

// InitManager initializes the global context manager
func InitManager(defaultCwd string) {
	globalManager = &Manager{
		contexts:   make(map[string]*Context),
		defaultCwd: defaultCwd,
	}
}

// CreateContext creates a new interpreter context
func (m *Manager) CreateContext(logger *slog.Logger, id, cwd, language string) (*Context, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.contexts[id]; exists {
		return nil, fmt.Errorf("context with ID '%s' already exists", id)
	}

	if language != "" && language != LanguagePython {
		return nil, fmt.Errorf("unsupported language: %s (only '%s' is supported)", language, LanguagePython)
	}
	if language == "" {
		language = LanguagePython
	}

	if cwd == "" {
		cwd = m.defaultCwd
	}

	iCtx := &Context{
		info: ContextInfo{
			ID:        id,
			Cwd:       cwd,
			CreatedAt: time.Now(),
			Active:    false,
			Language:  language,
		},
		logger: logger.With(slog.String("context_id", id)),
	}

	err := iCtx.start()
	if err != nil {
		return nil, fmt.Errorf("failed to start context: %w", err)
	}

	m.contexts[id] = iCtx
	return iCtx, nil
}

// GetContext retrieves an existing context by ID
func (m *Manager) GetContext(id string) (*Context, error) {
	m.mu.RLock()
	iCtx, exists := m.contexts[id]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("context with ID '%s' not found", id)
	}

	info := iCtx.Info()
	if !info.Active {
		err := iCtx.start()
		if err != nil {
			return nil, fmt.Errorf("failed to restart context '%s': %w", id, err)
		}
	}

	return iCtx, nil
}

// GetOrCreateDefaultContext gets or creates the default context
func (m *Manager) GetOrCreateDefaultContext(logger *slog.Logger) (*Context, error) {
	m.mu.RLock()
	iCtx, exists := m.contexts["default"]
	m.mu.RUnlock()

	if exists {
		info := iCtx.Info()
		if !info.Active {
			err := iCtx.start()
			if err != nil {
				return nil, fmt.Errorf("failed to restart default context: %w", err)
			}
		}
		return iCtx, nil
	}

	return m.CreateContext(logger, "default", m.defaultCwd, LanguagePython)
}

// DeleteContext removes a context and shuts it down
func (m *Manager) DeleteContext(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	iCtx, exists := m.contexts[id]
	if !exists {
		return common_errors.NewNotFoundError(fmt.Errorf("context with ID '%s' not found", id))
	}

	iCtx.shutdown()
	delete(m.contexts, id)
	return nil
}

// ListContexts returns information about all contexts
func (m *Manager) ListContexts() []ContextInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	contexts := make([]ContextInfo, 0, len(m.contexts))
	for _, iCtx := range m.contexts {
		contexts = append(contexts, iCtx.Info())
	}
	return contexts
}

// Global convenience functions

// CreateContext creates a new context using the global manager
func CreateContext(logger *slog.Logger, cwd, language string) (*Context, error) {
	if globalManager == nil {
		return nil, fmt.Errorf("context manager not initialized")
	}
	id := uuid.NewString()
	return globalManager.CreateContext(logger, id, cwd, language)
}

// GetContext gets a context by ID using the global manager
func GetContext(id string) (*Context, error) {
	if globalManager == nil {
		return nil, fmt.Errorf("context manager not initialized")
	}
	return globalManager.GetContext(id)
}

// GetOrCreateDefaultContext gets or creates the default context
func GetOrCreateDefaultContext(logger *slog.Logger) (*Context, error) {
	if globalManager == nil {
		return nil, fmt.Errorf("context manager not initialized")
	}
	return globalManager.GetOrCreateDefaultContext(logger)
}

// DeleteContext deletes a context using the global manager
func DeleteContext(id string) error {
	if globalManager == nil {
		return fmt.Errorf("context manager not initialized")
	}
	return globalManager.DeleteContext(id)
}

// ListContexts lists all contexts using the global manager
func ListContexts() []ContextInfo {
	if globalManager == nil {
		return []ContextInfo{}
	}
	return globalManager.ListContexts()
}
