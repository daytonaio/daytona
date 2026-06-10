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
	var publishes atomic.Int32
	countPublish := func(computeruse.IComputerUse) { publishes.Add(1) }
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
		firstImpl, firstErr = getOrSpawn(spawn, countPublish)
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
			results[i], errs[i] = getOrSpawn(spawn, countPublish)
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
	if got := publishes.Load(); got != 1 {
		t.Fatalf("publish ran %d times, want exactly 1 (only the fresh spawn publishes)", got)
	}
}

func TestGetOrSpawnDoesNotCacheFailure(t *testing.T) {
	resetPluginRef()
	t.Cleanup(resetPluginRef)

	var spawns atomic.Int32
	boom := errors.New("boom")
	impl := &fakeComputerUse{}

	var published []computeruse.IComputerUse
	record := func(impl computeruse.IComputerUse) { published = append(published, impl) }

	failThenSucceed := func() (*plugin.Client, computeruse.IComputerUse, error) {
		if spawns.Add(1) == 1 {
			return nil, nil, boom
		}
		return nil, impl, nil
	}

	if _, err := getOrSpawn(failThenSucceed, record); !errors.Is(err, boom) {
		t.Fatalf("got %v, want spawn error", err)
	}
	if cachedImpl() != nil {
		t.Fatal("failed spawn must not cache an impl")
	}
	if len(published) != 0 {
		t.Fatalf("failed spawn must not publish, got %v", published)
	}

	if got, err := getOrSpawn(failThenSucceed, record); err != nil || got != impl {
		t.Fatalf("retry after failure: got (%v, %v), want spawned impl", got, err)
	}
	if got, err := getOrSpawn(failThenSucceed, record); err != nil || got != impl {
		t.Fatalf("cached call: got (%v, %v), want cached impl", got, err)
	}
	if got := spawns.Load(); got != 2 {
		t.Fatalf("spawn ran %d times, want 2 (one failure, one success)", got)
	}
	if len(published) != 1 || published[0] != impl {
		t.Fatalf("want exactly one publish of the spawned impl, got %v", published)
	}
}

func TestKillComputerUseWaitsForInflightSpawn(t *testing.T) {
	resetPluginRef()
	t.Cleanup(resetPluginRef)

	impl := &fakeComputerUse{}
	lazy := computeruse.NewLazyComputerUse()
	entered := make(chan struct{})
	release := make(chan struct{})

	spawnDone := make(chan struct{})
	go func() {
		defer close(spawnDone)
		got, err := getOrSpawn(func() (*plugin.Client, computeruse.IComputerUse, error) {
			close(entered)
			<-release
			return nil, impl, nil
		}, lazy.Set)
		if err != nil || got != impl {
			t.Errorf("spawn caller: got (%v, %v), want spawned impl", got, err)
		}
	}()
	<-entered

	killDone := make(chan struct{})
	go func() {
		defer close(killDone)
		KillComputerUse(lazy.Set)
	}()

	select {
	case <-killDone:
		t.Fatal("KillComputerUse returned while a spawn was in flight; it must wait for the manager lock")
	case <-time.After(50 * time.Millisecond):
	}
	if lazy.IsReady() {
		t.Fatal("publish must not happen before the spawn completes")
	}

	close(release)
	<-spawnDone
	<-killDone

	// getOrSpawn caches and publishes under the lock before kill can acquire
	// it, so the kill always observes — and clears — the freshly spawned
	// instance, in both the manager cache and the published external cache.
	if cachedImpl() != nil {
		t.Fatal("kill racing a spawn must clear the freshly cached impl")
	}
	if lazy.IsReady() {
		t.Fatal("kill racing a spawn must clear the published external cache")
	}
}

// TestKillComputerUseWithoutSpawnPublishesClear pins the shutdown-path
// contract: KillComputerUse is safe to call unconditionally — with nothing
// spawned it kills nothing but still publishes a nil impl.
func TestKillComputerUseWithoutSpawnPublishesClear(t *testing.T) {
	resetPluginRef()
	t.Cleanup(resetPluginRef)

	var published []computeruse.IComputerUse
	KillComputerUse(func(impl computeruse.IComputerUse) { published = append(published, impl) })

	if len(published) != 1 || published[0] != nil {
		t.Fatalf("kill with nothing spawned must still publish nil exactly once, got %v", published)
	}
	if cachedImpl() != nil {
		t.Fatal("manager cache must stay empty")
	}
}
