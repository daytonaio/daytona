// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

const subscriptionTTL = 5 * time.Minute

type managedSubscription struct {
	unsubscribe func()
	timer       *time.Timer
}

type EventSubscriptionManager struct {
	mu            sync.Mutex
	dispatcher    *EventDispatcher
	subscriptions map[string]*managedSubscription
	closed        bool
}

func NewEventSubscriptionManager(dispatcher *EventDispatcher) *EventSubscriptionManager {
	return &EventSubscriptionManager{
		dispatcher:    dispatcher,
		subscriptions: make(map[string]*managedSubscription),
	}
}

func (m *EventSubscriptionManager) Subscribe(resourceID string, handler SandboxEventHandler, events []string) string {
	if m == nil || m.dispatcher == nil {
		return ""
	}
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		// Reject after shutdown to prevent use-after-close
		return ""
	}
	m.mu.Unlock()

	subID := uuid.NewString()
	unsubscribe := m.dispatcher.Subscribe(resourceID, handler, events)
	subscription := &managedSubscription{unsubscribe: unsubscribe}

	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		if unsubscribe != nil {
			unsubscribe()
		}
		// Reject after shutdown to prevent use-after-close
		return ""
	}
	m.subscriptions[subID] = subscription
	m.resetTimerLocked(subID, subscription)
	m.mu.Unlock()

	return subID
}

func (m *EventSubscriptionManager) Refresh(subID string) bool {
	if m == nil {
		return false
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		// Reject after shutdown to prevent use-after-close
		return false
	}

	subscription, ok := m.subscriptions[subID]
	if !ok {
		return false
	}

	m.resetTimerLocked(subID, subscription)
	return true
}

func (m *EventSubscriptionManager) Unsubscribe(subID string) {
	if m == nil {
		return
	}

	var unsubscribe func()

	m.mu.Lock()
	subscription, ok := m.subscriptions[subID]
	if ok {
		delete(m.subscriptions, subID)
		if subscription.timer != nil {
			subscription.timer.Stop()
		}
		unsubscribe = subscription.unsubscribe
	}
	m.mu.Unlock()

	if unsubscribe != nil {
		unsubscribe()
	}
}

func (m *EventSubscriptionManager) Shutdown() {
	if m == nil {
		return
	}

	var subscriptions []*managedSubscription

	m.mu.Lock()
	m.closed = true
	for _, subscription := range m.subscriptions {
		subscriptions = append(subscriptions, subscription)
	}
	m.subscriptions = make(map[string]*managedSubscription)
	m.mu.Unlock()

	for _, subscription := range subscriptions {
		if subscription.timer != nil {
			subscription.timer.Stop()
		}
		if subscription.unsubscribe != nil {
			subscription.unsubscribe()
		}
	}
}

func (m *EventSubscriptionManager) resetTimerLocked(subID string, subscription *managedSubscription) {
	if subscription.timer != nil {
		subscription.timer.Stop()
	}

	var timer *time.Timer
	timer = time.AfterFunc(subscriptionTTL, func() {
		var unsubscribe func()

		m.mu.Lock()
		current, ok := m.subscriptions[subID]
		if ok && current == subscription && current.timer == timer {
			delete(m.subscriptions, subID)
			unsubscribe = current.unsubscribe
		}
		m.mu.Unlock()

		if unsubscribe != nil {
			unsubscribe()
		}
	})

	subscription.timer = timer
}
