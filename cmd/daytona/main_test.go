package main

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/daytonaio/daytona/internal/listview"
)

func TestCheckComponentHealth(t *testing.T) {
	origAPIAddr := apiServerAddr
	origHeadscaleAddr := headscaleAddr
	origRegistryAddr := registryAddr
	origMaxRetries := maxRetries
	origRetryDelay := retryDelay

	defer func() {
		apiServerAddr = origAPIAddr
		headscaleAddr = origHeadscaleAddr
		registryAddr = origRegistryAddr
		maxRetries = origMaxRetries
		retryDelay = origRetryDelay
	}()

	maxRetries = 3
	retryDelay = 100 * time.Millisecond

	// Mock API server
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/health" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer apiServer.Close()
	apiServerAddr = apiServer.URL

	// Mock Headscale server
	headscaleServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer headscaleServer.Close()
	headscaleAddr = headscaleServer.URL

	// Setup mock registry
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create test listener: %v", err)
	}
	defer listener.Close()
	registryAddr = listener.Addr().String()

	err = checkComponentHealth(2 * time.Second)
	if err != nil {
		t.Errorf("checkComponentHealth failed: %v", err)
	}
}

func TestFailedAPIServer(t *testing.T) {
	origAPIAddr := apiServerAddr
	origMaxRetries := maxRetries
	origRetryDelay := retryDelay

	defer func() {
		apiServerAddr = origAPIAddr
		maxRetries = origMaxRetries
		retryDelay = origRetryDelay
	}()

	maxRetries = 2
	retryDelay = 100 * time.Millisecond

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	apiServerAddr = server.URL

	err := checkComponentHealth(1 * time.Second)
	if err == nil {
		t.Error("Expected error for failed API server, got nil")
	}
}

func TestFailedHeadscaleServer(t *testing.T) {
	origHeadscaleAddr := headscaleAddr
	origMaxRetries := maxRetries
	origRetryDelay := retryDelay

	defer func() {
		headscaleAddr = origHeadscaleAddr
		maxRetries = origMaxRetries
		retryDelay = origRetryDelay
	}()

	maxRetries = 2
	retryDelay = 100 * time.Millisecond

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	headscaleAddr = server.URL

	err := checkComponentHealth(1 * time.Second)
	if err == nil {
		t.Error("Expected error for failed headscale server, got nil")
	}
}

func TestContextTimeout(t *testing.T) {
	origAPIAddr := apiServerAddr
	origTimeout := defaultTimeout
	origMaxRetries := maxRetries
	origRetryDelay := retryDelay

	defer func() {
		apiServerAddr = origAPIAddr
		defaultTimeout = origTimeout
		maxRetries = origMaxRetries
		retryDelay = origRetryDelay
	}()

	maxRetries = 2
	retryDelay = 100 * time.Millisecond
	defaultTimeout = 1 * time.Millisecond
	apiServerAddr = "http://localhost:0"

	err := checkComponentHealth(50 * time.Millisecond)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}

func TestProviderCheck(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err := checkProviders(ctx)
	if err != nil {
		t.Errorf("Provider check failed: %v", err)
	}
}

func TestLocalRegistryCheck(t *testing.T) {
	origRegistryAddr := registryAddr
	defer func() { registryAddr = origRegistryAddr }()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create test listener: %v", err)
	}
	defer listener.Close()
	registryAddr = listener.Addr().String()

	err = checkLocalRegistry(context.Background())
	if err != nil {
		t.Errorf("Local registry check failed: %v", err)
	}
}

func TestRetryBehavior(t *testing.T) {
	origAPIAddr := apiServerAddr
	origRegistryAddr := registryAddr
	origMaxRetries := maxRetries
	origRetryDelay := retryDelay

	defer func() {
		apiServerAddr = origAPIAddr
		registryAddr = origRegistryAddr
		maxRetries = origMaxRetries
		retryDelay = origRetryDelay
	}()

	maxRetries = 3
	retryDelay = 100 * time.Millisecond

	// Setup API server with retry behavior
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	apiServerAddr = server.URL

	// Setup mock registry
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create test listener: %v", err)
	}
	defer listener.Close()
	registryAddr = listener.Addr().String()

	err = checkComponentHealth(2 * time.Second)
	if err != nil {
		t.Errorf("Expected success after retries, got error: %v", err)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestListView(t *testing.T) {
	lv := listview.New()
	if lv == nil {
		t.Error("Failed to create ListView")
	}

}
