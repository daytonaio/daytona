// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"sync"

	"github.com/tidwall/hashmap"
)

type Tracker[K comparable] struct {
	mu  sync.RWMutex
	set hashmap.Set[K]
}

func (t *Tracker[K]) Add(entry K) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.set.Insert(entry)
}

func (t *Tracker[K]) Remove(entry K) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.set.Delete(entry)
}

func (t *Tracker[K]) Contains(entry K) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.set.Contains(entry)
}
