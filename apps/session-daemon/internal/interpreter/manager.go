// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package interpreter

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/daytonaio/session-daemon/internal/config"
)

var (
	ErrContextExists   = errors.New("context already exists")
	ErrContextNotFound = errors.New("context not found")
	ErrCapacity        = errors.New("context capacity exceeded")
	ErrUnsupportedLang = errors.New("unsupported language")
)

// Manager owns the registry of contexts and the per-language Worker factories.
// It is concurrency-safe and runs an idle-context sweeper goroutine.
type Manager struct {
	cfg    *config.Config
	logger *slog.Logger

	pyFactory   WorkerFactory
	tsFactory   WorkerFactory
	bashFactory WorkerFactory

	mu       sync.RWMutex
	contexts map[string]*Session

	stopCh chan struct{}
	wg     sync.WaitGroup
}

// WorkerFactory produces a Worker for a context, given the context id and a
// handler that receives chunks tagged with that context id.
type WorkerFactory interface {
	Create(ctxID string, req CreateSessionRequest, onChunk func(*WorkerChunk)) (Worker, error)
	ListPackages() ([]PackageInfo, error)
	Shutdown()
}

// NewManager constructs a Manager and starts background goroutines.
func NewManager(cfg *config.Config, logger *slog.Logger, pyFactory, tsFactory, bashFactory WorkerFactory) *Manager {
	m := &Manager{
		cfg:         cfg,
		logger:      logger.With(slog.String("component", "manager")),
		pyFactory:   pyFactory,
		tsFactory:   tsFactory,
		bashFactory: bashFactory,
		contexts:    make(map[string]*Session),
		stopCh:      make(chan struct{}),
	}

	m.wg.Add(1)
	go m.idleSweeperLoop()

	if cfg.ApiIdleTtlHintSeconds > 0 {
		minDaemonIdle := cfg.ApiIdleTtlHintSeconds * 3 / 2 // 1.5x
		if cfg.ContextIdleTTLSeconds < minDaemonIdle {
			m.logger.Warn(
				"daemon idle TTL is shorter than 1.5x the API idle TTL hint; contexts may be reaped before the API expects",
				slog.Int("daemon_idle_ttl_seconds", cfg.ContextIdleTTLSeconds),
				slog.Int("api_idle_ttl_seconds_hint", cfg.ApiIdleTtlHintSeconds),
				slog.Int("recommended_min_daemon_idle_ttl_seconds", minDaemonIdle),
			)
		}
	}

	return m
}

// Close stops the sweeper, shuts down every context, and tears down the engine factories.
func (m *Manager) Close() {
	close(m.stopCh)
	m.wg.Wait()

	m.mu.Lock()
	for id, c := range m.contexts {
		c.shutdown()
		delete(m.contexts, id)
	}
	m.mu.Unlock()

	if m.pyFactory != nil {
		m.pyFactory.Shutdown()
	}
	if m.tsFactory != nil {
		m.tsFactory.Shutdown()
	}
	if m.bashFactory != nil {
		m.bashFactory.Shutdown()
	}
}

// CreateSession registers a new context with the daemon. id is supplied by the API
// and reused as the context's identity end-to-end (one id, one row, one identity).
func (m *Manager) CreateSession(req CreateSessionRequest) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.contexts[req.ID]; exists {
		return nil, ErrContextExists
	}

	factory, err := m.factoryFor(req.Language)
	if err != nil {
		return nil, err
	}

	if err := m.checkCapacityLocked(req.Language); err != nil {
		return nil, err
	}

	cwd := req.Cwd
	if cwd == "" {
		cwd = m.cfg.WorkspaceRoot
	}

	now := time.Now()
	c := &Session{
		info: SessionInfo{
			ID:         req.ID,
			Language:   normalizeLanguage(req.Language),
			Cwd:        cwd,
			CreatedAt:  now,
			LastUsedAt: now,
			Active:     true,
		},
		logger: m.logger,
	}

	worker, err := factory.Create(req.ID, req, c.handleChunk)
	if err != nil {
		return nil, fmt.Errorf("worker create: %w", err)
	}
	c.worker = worker
	c.startQueue()

	m.contexts[req.ID] = c
	m.logger.Debug("context created",
		slog.String("id", req.ID),
		slog.String("language", c.info.Language),
		slog.Int("memoryLimitMb", req.MemoryLimitMB),
	)
	return c, nil
}

func (m *Manager) factoryFor(lang string) (WorkerFactory, error) {
	switch normalizeLanguage(lang) {
	case LanguagePython:
		if m.pyFactory == nil {
			return nil, fmt.Errorf("%w: python factory not registered", ErrUnsupportedLang)
		}
		return m.pyFactory, nil
	case LanguageTypeScript:
		if m.tsFactory == nil {
			return nil, fmt.Errorf("%w: typescript factory not registered", ErrUnsupportedLang)
		}
		return m.tsFactory, nil
	case LanguageBash:
		if m.bashFactory == nil {
			return nil, fmt.Errorf("%w: bash factory not registered", ErrUnsupportedLang)
		}
		return m.bashFactory, nil
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnsupportedLang, lang)
	}
}

func (m *Manager) checkCapacityLocked(lang string) error {
	pyCount, tsCount, bashCount := 0, 0, 0
	for _, c := range m.contexts {
		switch c.info.Language {
		case LanguagePython:
			pyCount++
		case LanguageTypeScript:
			tsCount++
		case LanguageBash:
			bashCount++
		}
	}
	switch normalizeLanguage(lang) {
	case LanguagePython:
		if pyCount >= m.cfg.PyMaxContexts {
			return fmt.Errorf("%w: python contexts at cap (%d)", ErrCapacity, m.cfg.PyMaxContexts)
		}
	case LanguageTypeScript:
		if tsCount >= m.cfg.TSMaxContexts {
			return fmt.Errorf("%w: typescript contexts at cap (%d)", ErrCapacity, m.cfg.TSMaxContexts)
		}
	case LanguageBash:
		if bashCount >= m.cfg.BashMaxContexts {
			return fmt.Errorf("%w: bash contexts at cap (%d)", ErrCapacity, m.cfg.BashMaxContexts)
		}
	}
	return nil
}

// GetSession returns the context with the given id.
func (m *Manager) GetSession(id string) (*Session, error) {
	m.mu.RLock()
	c, ok := m.contexts[id]
	m.mu.RUnlock()
	if !ok {
		return nil, ErrContextNotFound
	}
	return c, nil
}

// DeleteSession shuts the context down and removes it from the registry.
func (m *Manager) DeleteSession(id string) error {
	m.mu.Lock()
	c, ok := m.contexts[id]
	if !ok {
		m.mu.Unlock()
		return ErrContextNotFound
	}
	delete(m.contexts, id)
	m.mu.Unlock()

	c.shutdown()
	m.logger.Debug("context deleted", slog.String("id", id))
	return nil
}

// deleteIfIdle removes a context only if it has no exec queued or running, re-checking
// under the write lock to close the TOCTOU race with the idle sweeper: a context that
// was idle when the sweep snapshot was taken can accept a job before deletion, and that
// job must not be torn down. Used only by sweepIdle; explicit DeleteSession is
// unconditional by design.
func (m *Manager) deleteIfIdle(id string) {
	m.mu.Lock()
	c, ok := m.contexts[id]
	if !ok {
		m.mu.Unlock()
		return
	}
	if c.hasInflightWork() {
		m.mu.Unlock()
		return
	}
	delete(m.contexts, id)
	m.mu.Unlock()

	c.shutdown()
	m.logger.Debug("context deleted (idle)", slog.String("id", id))
}

// ListSessions returns a snapshot of every active context's info.
func (m *Manager) ListSessions() []SessionInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]SessionInfo, 0, len(m.contexts))
	for _, c := range m.contexts {
		out = append(out, c.snapshotInfo())
	}
	return out
}

// ListPackages returns the curated package catalog for the given language.
func (m *Manager) ListPackages(language string) ([]PackageInfo, error) {
	factory, err := m.factoryFor(language)
	if err != nil {
		return nil, err
	}
	return factory.ListPackages()
}

// Healthz reports daemon readiness.
func (m *Manager) Healthz() bool {
	return true
}

// LoadCounts returns a snapshot of context concurrency for the /load endpoint:
// total active contexts, contexts with an in-flight exec, and the per-language caps.
func (m *Manager) LoadCounts() (active, busy, pyMax, tsMax, bashMax int) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	active = len(m.contexts)
	for _, c := range m.contexts {
		if c.IsBusy() {
			busy++
		}
	}
	return active, busy, m.cfg.PyMaxContexts, m.cfg.TSMaxContexts, m.cfg.BashMaxContexts
}

// WorkspaceRoot exposes the configured workspace root so the server can statfs the
// sandbox's own volume for the /load disk metric.
func (m *Manager) WorkspaceRoot() string { return m.cfg.WorkspaceRoot }

func normalizeLanguage(lang string) string {
	switch lang {
	case LanguageJavaScript:
		return LanguageTypeScript
	case "sh":
		return LanguageBash
	case "":
		return LanguagePython
	default:
		return lang
	}
}

// idleSweeperLoop runs every cfg.ContextGCIntervalSec and disposes any context
// whose lastUsedAt is older than cfg.ContextIdleTTLSeconds.
func (m *Manager) idleSweeperLoop() {
	defer m.wg.Done()

	if m.cfg.ContextIdleTTLSeconds <= 0 || m.cfg.ContextGCIntervalSec <= 0 {
		return
	}
	ticker := time.NewTicker(time.Duration(m.cfg.ContextGCIntervalSec) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.sweepIdle()
		}
	}
}

func (m *Manager) sweepIdle() {
	threshold := time.Now().Add(-time.Duration(m.cfg.ContextIdleTTLSeconds) * time.Second)
	stale := make([]string, 0)

	m.mu.RLock()
	for id, c := range m.contexts {
		if c.snapshotInfo().LastUsedAt.Before(threshold) && !c.hasInflightWork() {
			stale = append(stale, id)
		}
	}
	m.mu.RUnlock()

	for _, id := range stale {
		m.logger.Debug("sweeping idle context", slog.String("id", id))
		m.deleteIfIdle(id)
	}
}
