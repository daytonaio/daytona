// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package mock

import (
	"context"
	"io"
	"sync"

	"github.com/daytonaio/mock-runner/pkg/cache"
	"github.com/daytonaio/mock-runner/pkg/toolbox"
)

// MockClientConfig contains configuration for the mock client
type MockClientConfig struct {
	StatesCache       *cache.StatesCache
	LogWriter         io.Writer
	ToolboxContainer  *toolbox.ToolboxContainer
}

// MockClient implements sandbox and image operations without real Docker containers
type MockClient struct {
	statesCache      *cache.StatesCache
	logWriter        io.Writer
	toolboxContainer *toolbox.ToolboxContainer

	// In-memory storage for sandboxes
	sandboxes     map[string]*SandboxInfo
	sandboxesMu   sync.RWMutex

	// In-memory storage for images
	images   map[string]*ImageInfo
	imagesMu sync.RWMutex
}

// SandboxInfo holds mock sandbox state
type SandboxInfo struct {
	ID               string
	UserId           string
	Snapshot         string
	OsUser           string
	CpuQuota         int64
	GpuQuota         int64
	MemoryQuota      int64
	StorageQuota     int64
	Env              map[string]string
	Metadata         map[string]string
	NetworkBlockAll  *bool
	NetworkAllowList *string
}

// ImageInfo holds mock image information
type ImageInfo struct {
	Name       string
	Tag        string
	Size       int64
	Entrypoint []string
	Cmd        []string
	Hash       string
}

// NewMockClient creates a new mock client
func NewMockClient(config MockClientConfig) *MockClient {
	return &MockClient{
		statesCache:      config.StatesCache,
		logWriter:        config.LogWriter,
		toolboxContainer: config.ToolboxContainer,
		sandboxes:        make(map[string]*SandboxInfo),
		images:           make(map[string]*ImageInfo),
	}
}

// GetToolboxContainerIP returns the IP of the shared toolbox container
func (m *MockClient) GetToolboxContainerIP() string {
	if m.toolboxContainer != nil {
		return m.toolboxContainer.GetIP()
	}
	return ""
}

// GetToolboxContainer returns the toolbox container manager
func (m *MockClient) GetToolboxContainer() *toolbox.ToolboxContainer {
	return m.toolboxContainer
}

// getSandbox returns sandbox info by ID
func (m *MockClient) getSandbox(id string) (*SandboxInfo, bool) {
	m.sandboxesMu.RLock()
	defer m.sandboxesMu.RUnlock()
	sandbox, ok := m.sandboxes[id]
	return sandbox, ok
}

// setSandbox stores sandbox info
func (m *MockClient) setSandbox(sandbox *SandboxInfo) {
	m.sandboxesMu.Lock()
	defer m.sandboxesMu.Unlock()
	m.sandboxes[sandbox.ID] = sandbox
}

// deleteSandbox removes sandbox info
func (m *MockClient) deleteSandbox(id string) {
	m.sandboxesMu.Lock()
	defer m.sandboxesMu.Unlock()
	delete(m.sandboxes, id)
}

// getImage returns image info by name
func (m *MockClient) getImage(name string) (*ImageInfo, bool) {
	m.imagesMu.RLock()
	defer m.imagesMu.RUnlock()
	img, ok := m.images[name]
	return img, ok
}

// setImage stores image info
func (m *MockClient) setImage(img *ImageInfo) {
	m.imagesMu.Lock()
	defer m.imagesMu.Unlock()
	m.images[img.Name] = img
}

// deleteImage removes image info
func (m *MockClient) deleteImage(name string) {
	m.imagesMu.Lock()
	defer m.imagesMu.Unlock()
	delete(m.images, name)
}

// StatesCache returns the states cache
func (m *MockClient) StatesCache() *cache.StatesCache {
	return m.statesCache
}

// EnsureToolboxRunning ensures the toolbox container is running
func (m *MockClient) EnsureToolboxRunning(ctx context.Context) error {
	if m.toolboxContainer == nil {
		return nil
	}
	if !m.toolboxContainer.IsRunning() {
		return m.toolboxContainer.Start(ctx)
	}
	return nil
}



