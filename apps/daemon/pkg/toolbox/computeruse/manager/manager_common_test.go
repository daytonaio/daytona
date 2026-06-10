// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package manager

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"github.com/hashicorp/go-plugin"
)

// fakeComputerUse provides a unique non-nil IComputerUse identity for cache
// assertions. The embedded interface is never invoked.
type fakeComputerUse struct {
	computeruse.IComputerUse
}

func resetPluginRef() {
	computerUse.mu.Lock()
	defer computerUse.mu.Unlock()
	computerUse.client = nil
	computerUse.impl = nil
}

func cachedImpl() computeruse.IComputerUse {
	computerUse.mu.Lock()
	defer computerUse.mu.Unlock()
	return computerUse.impl
}

func TestGetOrSpawnSpawnsExactlyOnceUnderConcurrency(t *testing.T) {
	resetPluginRef()
	t.Cleanup(resetPluginRef)

	impl := &fakeComputerUse{}
	var spawns atomic.Int32
	entered := make(chan struct{})
	release := make(chan struct{})

	spawn := func() (*plugin.Client, computeruse.IComputerUse, error) {
		if spawns.Add(1) == 1 {
			close(entered)
		}
		<-release
		return nil, impl, nil
	}

	firstDone := make(chan struct{})
	var firstImpl computeruse.IComputerUse
	var firstErr error
	go func() {
		defer close(firstDone)
		firstImpl, firstErr = getOrSpawn(spawn)
	}()
	<-entered // first caller is inside spawn, holding the manager lock

	const others = 7
	var wg sync.WaitGroup
	results := make([]computeruse.IComputerUse, others)
	errs := make([]error, others)
	for i := 0; i < others; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i], errs[i] = getOrSpawn(spawn)
		}(i)
	}
	time.Sleep(20 * time.Millisecond) // let the others pile up on the lock
	close(release)
	<-firstDone
	wg.Wait()

	if firstErr != nil {
		t.Fatalf("first caller: unexpected error: %v", firstErr)
	}
	if firstImpl != impl {
		t.Fatalf("first caller: got %v, want the spawned impl", firstImpl)
	}
	for i := 0; i < others; i++ {
		if errs[i] != nil {
			t.Fatalf("caller %d: unexpected error: %v", i, errs[i])
		}
		if results[i] != impl {
			t.Fatalf("caller %d: got a different impl than the single spawn", i)
		}
	}
	if got := spawns.Load(); got != 1 {
		t.Fatalf("spawn ran %d times, want exactly 1", got)
	}
}

func TestGetOrSpawnDoesNotCacheFailure(t *testing.T) {
	resetPluginRef()
	t.Cleanup(resetPluginRef)

	var spawns atomic.Int32
	boom := errors.New("boom")
	impl := &fakeComputerUse{}

	failThenSucceed := func() (*plugin.Client, computeruse.IComputerUse, error) {
		if spawns.Add(1) == 1 {
			return nil, nil, boom
		}
		return nil, impl, nil
	}

	if _, err := getOrSpawn(failThenSucceed); !errors.Is(err, boom) {
		t.Fatalf("got %v, want spawn error", err)
	}
	if cachedImpl() != nil {
		t.Fatal("failed spawn must not cache an impl")
	}

	if got, err := getOrSpawn(failThenSucceed); err != nil || got != impl {
		t.Fatalf("retry after failure: got (%v, %v), want spawned impl", got, err)
	}
	if got, err := getOrSpawn(failThenSucceed); err != nil || got != impl {
		t.Fatalf("cached call: got (%v, %v), want cached impl", got, err)
	}
	if got := spawns.Load(); got != 2 {
		t.Fatalf("spawn ran %d times, want 2 (one failure, one success)", got)
	}
}

func TestKillComputerUseWaitsForInflightSpawn(t *testing.T) {
	resetPluginRef()
	t.Cleanup(resetPluginRef)

	impl := &fakeComputerUse{}
	entered := make(chan struct{})
	release := make(chan struct{})

	spawnDone := make(chan struct{})
	go func() {
		defer close(spawnDone)
		got, err := getOrSpawn(func() (*plugin.Client, computeruse.IComputerUse, error) {
			close(entered)
			<-release
			return nil, impl, nil
		})
		if err != nil || got != impl {
			t.Errorf("spawn caller: got (%v, %v), want spawned impl", got, err)
		}
	}()
	<-entered

	killDone := make(chan struct{})
	go func() {
		defer close(killDone)
		KillComputerUse()
	}()

	select {
	case <-killDone:
		t.Fatal("KillComputerUse returned while a spawn was in flight; it must wait for the manager lock")
	case <-time.After(50 * time.Millisecond):
	}

	close(release)
	<-spawnDone
	<-killDone

	// getOrSpawn caches under the lock before kill can acquire it, so the
	// kill always observes — and clears — the freshly spawned instance.
	if cachedImpl() != nil {
		t.Fatal("kill racing a spawn must clear the freshly cached impl")
	}
}
