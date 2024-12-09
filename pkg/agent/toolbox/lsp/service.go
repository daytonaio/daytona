package lsp

import (
	"fmt"
	"sync"
)

type LSPService struct {
	pathToProject string
	servers       map[string]LSPServer
}

var (
	instance *LSPService
	once     sync.Once
)

func GetLSPService() *LSPService {
	once.Do(func() {
		instance = &LSPService{
			pathToProject: "/home/daytona/learn-typescript",
			servers:       make(map[string]LSPServer),
		}
	})
	return instance
}

func (s *LSPService) Get(languageId string) (LSPServer, error) {
	if server, ok := s.servers[languageId]; ok {
		return server, nil
	}

	switch languageId {
	case "typescript":
		server := NewTypeScriptLSPServer()
		s.servers[languageId] = server
		return server, nil
	default:
		return nil, fmt.Errorf("unsupported language: %s", languageId)
	}
}

func (s *LSPService) Start(languageId string) error {
	server, ok := s.servers[languageId]
	if ok {
		if server.IsInitialized() {
			return nil
		}
	} else {
		newServer := NewTypeScriptLSPServer()
		s.servers[languageId] = newServer
		server = newServer
	}

	err := server.Initialize(s.pathToProject)
	if err != nil {
		return fmt.Errorf("failed to create TypeScript LSP server: %w", err)
	}

	return nil
}

func (s *LSPService) Shutdown(languageId string) error {
	server, ok := s.servers[languageId]
	if !ok {
		return fmt.Errorf("no server for language: %s", languageId)
	}
	server.Shutdown()

	delete(s.servers, languageId)
	return nil
}
