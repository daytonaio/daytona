// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package common

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

// EventSubscriber manages a Socket.IO connection and dispatches events.
// It is not sandbox-specific; events are dynamically registered and dispatched
// based on the event names passed to Subscribe.
type EventSubscriber struct {
	mu sync.RWMutex

	apiURL         string
	token          string
	organizationID string
	client         *socketIOClient
	connected      bool
	closed         bool
	failed         bool
	failError      string

	// listeners maps entity IDs to sets of handlers (keyed by subscription ID)
	listeners map[string]map[int]SandboxEventHandler
	nextSubID int

	// registeredEvents tracks which event names have been registered for dynamic dispatch
	registeredEvents map[string]bool

	// delayed disconnect state
	disconnectTimer *time.Timer
	disconnectGeneration uint64

	subscriptionTimers map[string]*time.Timer
	subscriptionTTLs   map[string]time.Duration

	// reconnection state
	connecting       bool
	reconnecting     bool
	reconnectAttempt int
	maxReconnects    int
	closeCh          chan struct{}
}

// NewEventSubscriber creates a new EventSubscriber.
func NewEventSubscriber(apiURL, token, organizationID string) *EventSubscriber {
	return &EventSubscriber{
		apiURL:           apiURL,
		token:            token,
		organizationID:   organizationID,
		listeners:        make(map[string]map[int]SandboxEventHandler),
		registeredEvents: make(map[string]bool),
		subscriptionTimers: make(map[string]*time.Timer),
		subscriptionTTLs:   make(map[string]time.Duration),
		maxReconnects:    100,
		closeCh:          make(chan struct{}),
	}
}

// EnsureConnected is idempotent: it ensures a connection attempt is in progress
// or already established. Non-blocking. Starts a background goroutine to
// connect if not already connected and no attempt is currently running.
func (es *EventSubscriber) EnsureConnected() {
	es.mu.Lock()
	if es.closed || es.connected || es.reconnecting || es.connecting {
		es.mu.Unlock()
		return
	}
	es.connecting = true
	es.mu.Unlock()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = es.Connect(ctx)
	}()
}

// Connect establishes the Socket.IO connection.
func (es *EventSubscriber) Connect(ctx context.Context) error {
	es.mu.Lock()
	if es.closed || es.connected {
		es.mu.Unlock()
		return nil
	}
	// Reinitialize closeCh if it was closed by a previous Disconnect
	select {
	case <-es.closeCh:
		es.closeCh = make(chan struct{})
	default:
	}
	es.mu.Unlock()

	es.mu.Lock()
	if es.client != nil {
		es.client.Close()
		es.client = nil
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
		es.connecting = false
		errMsg := es.failError
		es.mu.Unlock()
		return fmt.Errorf("%s", errMsg)
	}

	es.mu.Lock()
	if es.closed {
		es.connecting = false
		es.mu.Unlock()
		client.Close()
		return nil
	}
	es.client = client
	es.connected = true
	es.connecting = false
	es.failed = false
	es.failError = ""
	es.reconnectAttempt = 0
	es.mu.Unlock()

	return nil
}

// Subscribe registers a handler for events targeting a specific entity ID.
// The events parameter specifies which event names to listen for (registered idempotently).
// Returns an unsubscribe function.
const disconnectDelay = 30 * time.Second

func (es *EventSubscriber) Subscribe(sandboxID string, handler SandboxEventHandler, events []string, ttl time.Duration) func() {
	es.EnsureConnected()

	es.mu.Lock()
	es.disconnectGeneration++
	// Cancel any pending delayed disconnect
	if es.disconnectTimer != nil {
		es.disconnectTimer.Stop()
		es.disconnectTimer = nil
	}

	// Idempotently register requested events
	for _, event := range events {
		es.registeredEvents[event] = true
	}

	// Use a unique ID for stable removal
	subID := es.nextSubID
	es.nextSubID++
	if es.listeners[sandboxID] == nil {
		es.listeners[sandboxID] = make(map[int]SandboxEventHandler)
	}
	es.listeners[sandboxID][subID] = handler
	if ttl > 0 {
		es.subscriptionTTLs[sandboxID] = ttl
		es.startSubscriptionTimerLocked(sandboxID)
	} else {
		es.clearSubscriptionTimerLocked(sandboxID)
		delete(es.subscriptionTTLs, sandboxID)
	}
	es.mu.Unlock()

	return func() {
		es.mu.Lock()
		delete(es.listeners[sandboxID], subID)
		if len(es.listeners[sandboxID]) == 0 {
			es.unsubscribeResourceLocked(sandboxID)
		}

		// Schedule delayed disconnect when no entities are listening anymore
		if len(es.listeners) == 0 {
			es.scheduleDelayedDisconnectLocked()
		}
		es.mu.Unlock()
	}
}

func (es *EventSubscriber) RefreshSubscription(resourceID string) bool {
	es.mu.Lock()
	defer es.mu.Unlock()

	if _, ok := es.subscriptionTTLs[resourceID]; !ok {
		return false
	}

	es.startSubscriptionTimerLocked(resourceID)
	return true
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
	es.mu.Lock()
	defer es.mu.Unlock()
	es.disconnectLocked(true)
}

func (es *EventSubscriber) disconnect(permanent bool) {
	es.mu.Lock()
	defer es.mu.Unlock()
	es.disconnectLocked(permanent)
}

func (es *EventSubscriber) disconnectLocked(permanent bool) {
	select {
	case <-es.closeCh:
	default:
		close(es.closeCh)
	}
	if permanent {
		es.closed = true
	}
	es.connecting = false
	es.reconnecting = false
	if es.disconnectTimer != nil {
		es.disconnectTimer.Stop()
		es.disconnectTimer = nil
	}
	es.clearSubscriptionTimersLocked()
	if es.client != nil {
		es.client.Close()
		es.client = nil
	}
	es.connected = false
	es.listeners = make(map[string]map[int]SandboxEventHandler)
	es.registeredEvents = make(map[string]bool)
}

// handleEvent is called by the Socket.IO client for each event.
// It dynamically dispatches based on registered events rather than
// hardcoding specific event names.
func (es *EventSubscriber) handleEvent(eventName string, data json.RawMessage) {
	es.mu.RLock()
	registered := es.registeredEvents[eventName]
	es.mu.RUnlock()

	if !registered {
		return
	}

	// Extract the entity ID using a nested key lookup pattern:
	// try "sandbox", "volume", "snapshot", "runner" nested objects, then fall back to top-level "id".
	entityID := extractEntityID(data)
	if entityID == "" {
		return
	}

	// Build the SandboxEvent based on the event name
	var event SandboxEvent
	switch eventName {
	case "sandbox.state.updated":
		var stateEvent SandboxStateEvent
		if err := json.Unmarshal(data, &stateEvent); err != nil {
			return
		}
		event = SandboxEvent{
			Type:       SandboxEventStateUpdated,
			StateEvent: &stateEvent,
		}

	case "sandbox.desired-state.updated":
		var desiredStateEvent SandboxDesiredStateEvent
		if err := json.Unmarshal(data, &desiredStateEvent); err != nil {
			return
		}
		event = SandboxEvent{
			Type:              SandboxEventDesiredStateUpdated,
			DesiredStateEvent: &desiredStateEvent,
		}

	case "sandbox.created":
		var sandbox apiclient.Sandbox
		if err := json.Unmarshal(data, &sandbox); err != nil {
			return
		}
		event = SandboxEvent{
			Type:         SandboxEventCreated,
			CreatedEvent: &sandbox,
		}

	default:
		// For any other registered event, attempt to parse as a state event
		// and dispatch with the extracted entity ID
		var stateEvent SandboxStateEvent
		if err := json.Unmarshal(data, &stateEvent); err != nil {
			return
		}
		event = SandboxEvent{
			Type:       SandboxEventStateUpdated,
			StateEvent: &stateEvent,
		}
	}

	es.dispatch(entityID, event)
}

// extractEntityID extracts an entity ID from event data.
// It tries nested keys "sandbox", "volume", "snapshot", "runner" (looking for an "id" field
// inside them), then falls back to a top-level "id" field.
func extractEntityID(data json.RawMessage) string {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return ""
	}

	// Try nested entity keys
	nestedKeys := []string{"sandbox", "volume", "snapshot", "runner"}
	for _, key := range nestedKeys {
		if nested, ok := raw[key]; ok {
			var nestedObj map[string]json.RawMessage
			if err := json.Unmarshal(nested, &nestedObj); err == nil {
				if idRaw, ok := nestedObj["id"]; ok {
					var id string
					if err := json.Unmarshal(idRaw, &id); err == nil && id != "" {
						return id
					}
				}
			}
		}
	}

	// Fall back to top-level "id"
	if idRaw, ok := raw["id"]; ok {
		var id string
		if err := json.Unmarshal(idRaw, &id); err == nil {
			return id
		}
	}

	return ""
}

// dispatch sends an event to all handlers registered for an entity ID.
func (es *EventSubscriber) dispatch(entityID string, event SandboxEvent) {
	if entityID == "" {
		return
	}

	es.mu.RLock()
	handlerMap := es.listeners[entityID]
	handlers := make([]SandboxEventHandler, 0, len(handlerMap))
	for _, h := range handlerMap {
		handlers = append(handlers, h)
	}
	es.mu.RUnlock()

	for _, handler := range handlers {
		func() {
			defer func() {
				_ = recover() // Don't let a handler panic break other handlers
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

func (es *EventSubscriber) startSubscriptionTimerLocked(resourceID string) {
	ttl, ok := es.subscriptionTTLs[resourceID]
	if !ok || ttl <= 0 {
		return
	}

	es.clearSubscriptionTimerLocked(resourceID)

	var timer *time.Timer
	timer = time.AfterFunc(ttl, func() {
		es.mu.Lock()
		defer es.mu.Unlock()

		current, ok := es.subscriptionTimers[resourceID]
		if !ok || current != timer {
			return
		}

		delete(es.subscriptionTimers, resourceID)
		delete(es.subscriptionTTLs, resourceID)
		es.unsubscribeResourceLocked(resourceID)
		if len(es.listeners) == 0 {
			es.scheduleDelayedDisconnectLocked()
		}
	})

	es.subscriptionTimers[resourceID] = timer
}

func (es *EventSubscriber) clearSubscriptionTimerLocked(resourceID string) {
	if timer, ok := es.subscriptionTimers[resourceID]; ok {
		timer.Stop()
		delete(es.subscriptionTimers, resourceID)
	}
}

func (es *EventSubscriber) clearSubscriptionTimersLocked() {
	for resourceID, timer := range es.subscriptionTimers {
		timer.Stop()
		delete(es.subscriptionTimers, resourceID)
	}
	for resourceID := range es.subscriptionTTLs {
		delete(es.subscriptionTTLs, resourceID)
	}
}

func (es *EventSubscriber) unsubscribeResourceLocked(resourceID string) {
	delete(es.listeners, resourceID)
	es.clearSubscriptionTimerLocked(resourceID)
	delete(es.subscriptionTTLs, resourceID)
}

func (es *EventSubscriber) scheduleDelayedDisconnectLocked() {
	if es.disconnectTimer != nil {
		es.disconnectTimer.Stop()
	}
	myGeneration := es.disconnectGeneration

	es.disconnectTimer = time.AfterFunc(disconnectDelay, func() {
		es.mu.Lock()
		defer es.mu.Unlock()
		if myGeneration == es.disconnectGeneration && len(es.listeners) == 0 {
			es.disconnectLocked(false)
		}
	})
}
