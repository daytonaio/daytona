// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package interpreter

import (
	"context"
	"testing"
	"time"
)

// TestEnqueueAfterShutdownRejected verifies the shuttingDown gate: once shutdown
// has run, Enqueue refuses the job (returning a shutdown error) WITHOUT reserving
// an inflight slot. Before the fix, the select could still win the buffered send
// after the queue context was cancelled and the drainer returned, stranding the
// caller and leaking the inflight count that gates idle GC.
func TestEnqueueAfterShutdownRejected(t *testing.T) {
	c := &Session{info: SessionInfo{ID: "x", Language: LanguageBash}}
	c.startQueue()
	c.shutdown()

	select {
	case res := <-c.Enqueue("code", nil, 0, false):
		if res.Err == nil {
			t.Fatalf("expected a shutdown error from Enqueue after shutdown, got nil")
		}
	case <-time.After(time.Second):
		t.Fatal("Enqueue after shutdown blocked instead of returning a shutdown error")
	}

	if got := c.inflight.Load(); got != 0 {
		t.Fatalf("inflight leaked after a rejected enqueue: got %d, want 0", got)
	}
}

// TestDrainAndCloseAnswersBufferedJobs exercises drainAndClose directly: every
// job that was accepted (inflight reserved + buffered) before shutdown must get a
// shutdown error on its done channel and have its reservation released, so a
// caller can never block forever and idle GC is never wedged by a leaked count.
func TestDrainAndCloseAnswersBufferedJobs(t *testing.T) {
	c := &Session{info: SessionInfo{ID: "x"}}
	// Wire the queue WITHOUT a consumer goroutine so the jobs sit in the buffer,
	// exactly the state drainAndClose must mop up after processQueue has exited.
	c.queue = make(chan execJob, 8)
	c.queueCtx, c.queueStop = context.WithCancel(context.Background())

	d1 := make(chan execResult, 1)
	d2 := make(chan execResult, 1)
	c.inflight.Add(1)
	c.queue <- execJob{doneCh: d1}
	c.inflight.Add(1)
	c.queue <- execJob{doneCh: d2}

	c.shuttingDown = true
	c.queueStop()
	c.drainAndClose(c.queue)

	for i, d := range []chan execResult{d1, d2} {
		select {
		case r := <-d:
			if r.Err == nil {
				t.Fatalf("buffered job %d: expected a shutdown error, got nil", i)
			}
		default:
			t.Fatalf("buffered job %d was not answered by drainAndClose", i)
		}
	}
	if got := c.inflight.Load(); got != 0 {
		t.Fatalf("inflight not drained by drainAndClose: got %d, want 0", got)
	}
}
