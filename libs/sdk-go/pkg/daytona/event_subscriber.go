// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daytona

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
)

// SandboxStateEvent represents a sandbox.state.updated event payload.
type SandboxStateEvent struct {
	Sandbox  apiclient.Sandbox      `json:"sandbox"`
	OldState apiclient.SandboxState `json:"oldState"`
	NewState apiclient.SandboxState `json:"newState"`
}

// SandboxDesiredStateEvent represents a sandbox.desired-state.updated event payload.
type SandboxDesiredStateEvent struct {
	Sandbox         apiclient.Sandbox             `json:"sandbox"`
	OldDesiredState apiclient.SandboxDesiredState `json:"oldDesiredState"`
	NewDesiredState apiclient.SandboxDesiredState `json:"newDesiredState"`
}

// SandboxEventType identifies the type of sandbox event.
type SandboxEventType int

const (
	// SandboxEventStateUpdated is emitted when sandbox state changes.
	SandboxEventStateUpdated SandboxEventType = iota
	// SandboxEventDesiredStateUpdated is emitted when sandbox desired state changes.
	SandboxEventDesiredStateUpdated
	// SandboxEventCreated is emitted when a sandbox is created.
	SandboxEventCreated
)

// SandboxEvent wraps all possible sandbox event types.
type SandboxEvent struct {
	Type              SandboxEventType
	StateEvent        *SandboxStateEvent
	DesiredStateEvent *SandboxDesiredStateEvent
	CreatedEvent      *apiclient.Sandbox
}

// SandboxEventHandler is called when a sandbox event is received.
type SandboxEventHandler func(event SandboxEvent)

// EventSubscriber manages a Socket.IO connection and dispatches sandbox events.
type EventSubscriber struct {
	mu sync.RWMutex

	apiURL         string
	token          string
	organizationID string
	client         *socketIOClient
	connected      bool
	failed         bool
	failError      string

	// listeners maps sandbox IDs to sets of handlers
	listeners map[string][]SandboxEventHandler

	// delayed disconnect state
	disconnectTimer *time.Timer

	// reconnection state
	reconnecting     bool
	reconnectAttempt int
	maxReconnects    int
	closeCh          chan struct{}
}

// NewEventSubscriber creates a new EventSubscriber.
func NewEventSubscriber(apiURL, token, organizationID string) *EventSubscriber {
	return &EventSubscriber{
		apiURL:         apiURL,
		token:          token,
		organizationID: organizationID,
		listeners:      make(map[string][]SandboxEventHandler),
		maxReconnects:  10,
		closeCh:        make(chan struct{}),
	}
}

// Connect establishes the Socket.IO connection.
func (es *EventSubscriber) Connect(ctx context.Context) error {
	es.mu.Lock()
	if es.connected {
		es.mu.Unlock()
		return nil
	}
	es.mu.Unlock()

	timeout := 5 * time.Second
	if deadline, ok := ctx.Deadline(); ok {
		timeout = time.Until(deadline)
	}

	client, err := newSocketIOClient(socketIOClientConfig{
		APIURL:         es.apiURL,
		Token:          es.token,
		OrganizationID: es.organizationID,
		ConnectTimeout: timeout,
		EventHandler:   es.handleEvent,
		OnDisconnect:   es.handleDisconnect,
	})
	if err != nil {
		es.mu.Lock()
		es.failed = true
		es.failError = fmt.Sprintf("WebSocket connection failed: %v", err)
		es.mu.Unlock()
		return fmt.Errorf("%s", es.failError)
	}

	es.mu.Lock()
	es.client = client
	es.connected = true
	es.failed = false
	es.failError = ""
	es.reconnectAttempt = 0
	es.mu.Unlock()

	return nil
}

// Subscribe registers a handler for events targeting a specific sandbox.
// Returns an unsubscribe function.
const disconnectDelay = 30 * time.Second

func (es *EventSubscriber) Subscribe(sandboxID string, handler SandboxEventHandler) func() {
	es.mu.Lock()
	// Cancel any pending delayed disconnect
	if es.disconnectTimer != nil {
		es.disconnectTimer.Stop()
		es.disconnectTimer = nil
	}
	es.listeners[sandboxID] = append(es.listeners[sandboxID], handler)
	idx := len(es.listeners[sandboxID]) - 1
	es.mu.Unlock()

	return func() {
		es.mu.Lock()
		handlers := es.listeners[sandboxID]
		if idx < len(handlers) {
			handlers[idx] = handlers[len(handlers)-1]
			es.listeners[sandboxID] = handlers[:len(handlers)-1]
		}
		if len(es.listeners[sandboxID]) == 0 {
			delete(es.listeners, sandboxID)
		}

		// Schedule delayed disconnect when no sandboxes are listening anymore
		if len(es.listeners) == 0 {
			es.disconnectTimer = time.AfterFunc(disconnectDelay, func() {
				es.mu.RLock()
				empty := len(es.listeners) == 0
				es.mu.RUnlock()
				if empty {
					es.Disconnect()
				}
			})
		}
		es.mu.Unlock()
	}
}

// WaitForState waits for a sandbox to reach one of the target states.
// Returns the new state or an error if an error state is reached or context is cancelled.
func (es *EventSubscriber) WaitForState(
	ctx context.Context,
	sandboxID string,
	targetStates []apiclient.SandboxState,
	errorStates []apiclient.SandboxState,
) (apiclient.SandboxState, error) {
	es.mu.RLock()
	if es.failed {
		errMsg := es.failError
		es.mu.RUnlock()
		return "", fmt.Errorf("%s", errMsg)
	}
	es.mu.RUnlock()

	ch := make(chan SandboxEvent, 16)
	unsubscribe := es.Subscribe(sandboxID, func(event SandboxEvent) {
		select {
		case ch <- event:
		default:
			// Channel full, drop event (shouldn't happen with buffered channel)
		}
	})
	defer unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case event := <-ch:
			var newState apiclient.SandboxState
			if event.Type == SandboxEventStateUpdated && event.StateEvent != nil {
				newState = event.StateEvent.NewState
			}

			if newState == "" {
				continue
			}

			for _, target := range targetStates {
				if newState == target {
					return newState, nil
				}
			}
			for _, errState := range errorStates {
				if newState == errState {
					return newState, fmt.Errorf("sandbox entered error state: %s", newState)
				}
			}
		}
	}
}

// IsConnected returns whether the subscriber is currently connected.
func (es *EventSubscriber) IsConnected() bool {
	es.mu.RLock()
	defer es.mu.RUnlock()
	return es.connected
}

// IsFailed returns whether the subscriber has permanently failed.
func (es *EventSubscriber) IsFailed() bool {
	es.mu.RLock()
	defer es.mu.RUnlock()
	return es.failed
}

// FailError returns the failure error message, if any.
func (es *EventSubscriber) FailError() string {
	es.mu.RLock()
	defer es.mu.RUnlock()
	return es.failError
}

// Disconnect closes the connection and cleans up resources.
func (es *EventSubscriber) Disconnect() {
	select {
	case <-es.closeCh:
	default:
		close(es.closeCh)
	}

	es.mu.Lock()
	if es.client != nil {
		es.client.Close()
		es.client = nil
	}
	es.connected = false
	es.listeners = make(map[string][]SandboxEventHandler)
	es.mu.Unlock()
}

// handleEvent is called by the Socket.IO client for each event.
func (es *EventSubscriber) handleEvent(eventName string, data json.RawMessage) {
	switch eventName {
	case "sandbox.state.updated":
		var event SandboxStateEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return
		}
		es.dispatch(event.Sandbox.GetId(), SandboxEvent{
			Type:       SandboxEventStateUpdated,
			StateEvent: &event,
		})

	case "sandbox.desired-state.updated":
		var event SandboxDesiredStateEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return
		}
		es.dispatch(event.Sandbox.GetId(), SandboxEvent{
			Type:              SandboxEventDesiredStateUpdated,
			DesiredStateEvent: &event,
		})

	case "sandbox.created":
		var sandbox apiclient.Sandbox
		if err := json.Unmarshal(data, &sandbox); err != nil {
			return
		}
		es.dispatch(sandbox.GetId(), SandboxEvent{
			Type:         SandboxEventCreated,
			CreatedEvent: &sandbox,
		})
	}
}

// dispatch sends an event to all handlers registered for a sandbox ID.
func (es *EventSubscriber) dispatch(sandboxID string, event SandboxEvent) {
	if sandboxID == "" {
		return
	}

	es.mu.RLock()
	handlers := make([]SandboxEventHandler, len(es.listeners[sandboxID]))
	copy(handlers, es.listeners[sandboxID])
	es.mu.RUnlock()

	for _, handler := range handlers {
		func() {
			defer func() {
				recover() // Don't let a handler panic break other handlers
			}()
			handler(event)
		}()
	}
}

// handleDisconnect is called when the Socket.IO connection is lost.
func (es *EventSubscriber) handleDisconnect() {
	es.mu.Lock()
	es.connected = false

	// Check if we should reconnect
	select {
	case <-es.closeCh:
		// Intentionally closed, don't reconnect
		es.mu.Unlock()
		return
	default:
	}

	if es.reconnecting {
		es.mu.Unlock()
		return
	}
	es.reconnecting = true
	es.mu.Unlock()

	go es.reconnectLoop()
}

// reconnectLoop attempts to reconnect with exponential backoff.
func (es *EventSubscriber) reconnectLoop() {
	defer func() {
		es.mu.Lock()
		es.reconnecting = false
		es.mu.Unlock()
	}()

	for attempt := 0; attempt < es.maxReconnects; attempt++ {
		select {
		case <-es.closeCh:
			return
		default:
		}

		// Exponential backoff: 1s, 2s, 4s, 8s, ..., max 30s
		delay := time.Duration(1<<uint(attempt)) * time.Second
		if delay > 30*time.Second {
			delay = 30 * time.Second
		}

		select {
		case <-time.After(delay):
		case <-es.closeCh:
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := es.Connect(ctx)
		cancel()

		if err == nil {
			return
		}

		es.mu.Lock()
		es.reconnectAttempt = attempt + 1
		es.mu.Unlock()
	}

	// All reconnection attempts failed
	es.mu.Lock()
	es.failed = true
	es.failError = fmt.Sprintf("WebSocket reconnection failed after %d attempts", es.maxReconnects)
	es.mu.Unlock()
}
