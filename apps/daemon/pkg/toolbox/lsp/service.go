// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package lsp

import (
	"encoding/base64"
	"fmt"
	"sync"
)

type LSPService struct {
	servers map[string]LSPServer
}

var (
	instance *LSPService
	once     sync.Once
)

func GetLSPService() *LSPService {
	once.Do(func() {
		instance = &LSPService{
			servers: make(map[string]LSPServer),
		}
	})
	return instance
}

func (s *LSPService) Get(languageId string, pathToProject string) (LSPServer, error) {
	key := generateKey(languageId, pathToProject)

	if server, ok := s.servers[key]; ok {
		return server, nil
	}

	switch languageId {
	case "typescript":
		server := NewTypeScriptLSPServer()
		s.servers[key] = server
		return server, nil
	case "python":
		server := NewPythonLSPServer()
		s.servers[key] = server
		return server, nil
	default:
		return nil, fmt.Errorf("unsupported language: %s", languageId)
	}
}

func (s *LSPService) Start(languageId string, pathToProject string) error {
	key := generateKey(languageId, pathToProject)

	server, ok := s.servers[key]
	if ok {
		if server.IsInitialized() {
			return nil
		}
	} else {
		newServer := NewTypeScriptLSPServer()
		s.servers[key] = newServer
		server = newServer
	}

	err := server.Initialize(pathToProject)
	if err != nil {
		return fmt.Errorf("failed to create TypeScript LSP server: %w", err)
	}

	return nil
}

func (s *LSPService) Shutdown(languageId string, pathToProject string) error {
	key := generateKey(languageId, pathToProject)

	server, ok := s.servers[key]
	if !ok {
		return fmt.Errorf("no server for language: %s", languageId)
	}
	err := server.Shutdown()
	delete(s.servers, key)
	return err
}

func generateKey(languageId, pathToProject string) string {
	data := fmt.Sprintf("%s:%s", languageId, pathToProject)
	return base64.StdEncoding.EncodeToString([]byte(data))
}
