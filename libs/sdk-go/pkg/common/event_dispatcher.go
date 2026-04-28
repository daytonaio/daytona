// Copyright Daytona Platforms Inc.
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

type SandboxStateEvent struct {
	Sandbox  apiclient.Sandbox      `json:"sandbox"`
	OldState apiclient.SandboxState `json:"oldState"`
	NewState apiclient.SandboxState `json:"newState"`
}

type SandboxDesiredStateEvent struct {
	Sandbox         apiclient.Sandbox             `json:"sandbox"`
	OldDesiredState apiclient.SandboxDesiredState `json:"oldDesiredState"`
	NewDesiredState apiclient.SandboxDesiredState `json:"newDesiredState"`
}

type SandboxEventType int

const (
	SandboxEventStateUpdated SandboxEventType = iota
	SandboxEventDesiredStateUpdated
	SandboxEventCreated
)

type SandboxEvent struct {
	Type              SandboxEventType
	StateEvent        *SandboxStateEvent
	DesiredStateEvent *SandboxDesiredStateEvent
	CreatedEvent      *apiclient.Sandbox
}

type SandboxEventHandler func(event SandboxEvent)

type EventDispatcher struct {
	mu sync.RWMutex

	apiURL         string
	token          string
	organizationID string
	client         *socketIOClient
	clientGen      uint64
	connected      bool
	closed         bool
	failed         bool
	failError      string

	listeners map[string]map[int]SandboxEventHandler
	nextSubID int

	registeredEvents map[string]bool

	disconnectTimer      *time.Timer
	disconnectGeneration uint64

	connecting       bool
	reconnecting     bool
	reconnectAttempt int
	maxReconnects    int
	closeCh          chan struct{}
}

func NewEventDispatcher(apiURL, token, organizationID string) *EventDispatcher {
	return &EventDispatcher{
		apiURL:           apiURL,
		token:            token,
		organizationID:   organizationID,
		listeners:        make(map[string]map[int]SandboxEventHandler),
		registeredEvents: make(map[string]bool),
		maxReconnects:    100,
		closeCh:          make(chan struct{}),
	}
}

func (es *EventDispatcher) EnsureConnected() {
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

func (es *EventDispatcher) Connect(ctx context.Context) error {
	es.mu.Lock()
	if es.closed || es.connected {
		es.mu.Unlock()
		return nil
	}
	select {
	case <-es.closeCh:
		es.closeCh = make(chan struct{})
	default:
	}
	es.clientGen++
	clientGen := es.clientGen
	oldClient := es.client
	es.client = nil
	es.mu.Unlock()

	if oldClient != nil {
		oldClient.Close()
	}

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
		OnDisconnect: func() {
			es.handleDisconnect(clientGen)
		},
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
	if len(es.listeners) == 0 {
		es.scheduleDelayedDisconnectLocked()
	}
	es.mu.Unlock()

	return nil
}

func (es *EventDispatcher) RegisterEvents(events []string) {
	es.mu.Lock()
	defer es.mu.Unlock()

	for _, event := range events {
		es.registeredEvents[event] = true
	}
}

const disconnectDelay = 30 * time.Second

func (es *EventDispatcher) Subscribe(resourceID string, handler SandboxEventHandler, events []string) func() {
	es.mu.Lock()
	if es.closed {
		es.mu.Unlock()
		// No-op after disconnect
		return func() {}
	}
	es.disconnectGeneration++
	if es.disconnectTimer != nil {
		es.disconnectTimer.Stop()
		es.disconnectTimer = nil
	}

	subID := es.nextSubID
	es.nextSubID++
	if es.listeners[resourceID] == nil {
		es.listeners[resourceID] = make(map[int]SandboxEventHandler)
	}
	es.listeners[resourceID][subID] = handler
	// Mark events as registered under lock to prevent drops
	for _, event := range events {
		es.registeredEvents[event] = true
	}
	es.mu.Unlock()

	es.EnsureConnected()
	es.RegisterEvents(events)

	return func() {
		es.mu.Lock()
		delete(es.listeners[resourceID], subID)
		if len(es.listeners[resourceID]) == 0 {
			delete(es.listeners, resourceID)
		}

		if len(es.listeners) == 0 {
			es.scheduleDelayedDisconnectLocked()
		}
		es.mu.Unlock()
	}
}

func (es *EventDispatcher) IsConnected() bool {
	es.mu.RLock()
	defer es.mu.RUnlock()
	return es.connected
}

func (es *EventDispatcher) IsFailed() bool {
	es.mu.RLock()
	defer es.mu.RUnlock()
	return es.failed
}

func (es *EventDispatcher) FailError() string {
	es.mu.RLock()
	defer es.mu.RUnlock()
	return es.failError
}

func (es *EventDispatcher) Disconnect() {
	es.mu.Lock()
	defer es.mu.Unlock()
	es.disconnectLocked(true)
}

func (es *EventDispatcher) disconnectLocked(permanent bool) {
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
	if es.client != nil {
		es.client.Close()
		es.client = nil
	}
	es.connected = false
	es.listeners = make(map[string]map[int]SandboxEventHandler)
	es.registeredEvents = make(map[string]bool)
}

func (es *EventDispatcher) handleEvent(eventName string, data json.RawMessage) {
	es.mu.RLock()
	registered := es.registeredEvents[eventName]
	es.mu.RUnlock()

	if !registered {
		return
	}

	entityID := extractEntityID(data)
	if entityID == "" {
		return
	}

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
		return
	}

	es.dispatch(entityID, event)
}

func extractEntityID(data json.RawMessage) string {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return ""
	}

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

	if idRaw, ok := raw["id"]; ok {
		var id string
		if err := json.Unmarshal(idRaw, &id); err == nil {
			return id
		}
	}

	return ""
}

func (es *EventDispatcher) dispatch(entityID string, event SandboxEvent) {
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
				_ = recover()
			}()
			handler(event)
		}()
	}
}

func (es *EventDispatcher) handleDisconnect(clientGen uint64) {
	es.mu.Lock()
	if clientGen != es.clientGen {
		// Ignore disconnect from replaced client
		es.mu.Unlock()
		return
	}
	es.connected = false

	select {
	case <-es.closeCh:
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

func (es *EventDispatcher) reconnectLoop() {
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

	es.mu.Lock()
	es.failed = true
	es.failError = fmt.Sprintf("WebSocket reconnection failed after %d attempts", es.maxReconnects)
	es.mu.Unlock()
}

func (es *EventDispatcher) scheduleDelayedDisconnectLocked() {
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
