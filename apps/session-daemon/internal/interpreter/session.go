// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package interpreter

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// snapshotInfo returns a defensive copy of the context's info.
func (c *Session) snapshotInfo() SessionInfo {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.info
}

// SnapshotInfo is the public, read-only view of a context's metadata.
func (c *Session) SnapshotInfo() SessionInfo { return c.snapshotInfo() }

// startQueue spins up the FIFO worker goroutine on first use.
func (c *Session) startQueue() {
	c.queueOnce.Do(func() {
		c.queue = make(chan execJob, 128)
		c.queueCtx, c.queueStop = context.WithCancel(context.Background())
		go c.processQueue()
	})
}

// Enqueue queues a job on the context's FIFO and returns a channel that completes
// with the result (and any error). Callers MUST drain the result channel.
func (c *Session) Enqueue(code string, envs map[string]string, timeout time.Duration, reset bool) <-chan execResult {
	doneCh := make(chan execResult, 1)
	job := execJob{code: code, envs: envs, timeout: timeout, reset: reset, doneCh: doneCh}

	// Reserve the inflight slot, start the queue, and apply the shuttingDown gate
	// all under the one lock shutdown() uses (to set shuttingDown and read
	// c.queue/c.queueStop). This (1) lets a concurrent idle sweep observe the work,
	// (2) refuses the job once shutdown has begun — without the gate the select
	// below could pick the buffered send after drainAndClose returned, stranding the
	// job (caller blocks forever) and leaking the inflight count — and (3) keeps
	// startQueue INSIDE the gate, so a late Enqueue after shutdown can't spawn a
	// processQueue goroutine that shutdown already skipped (which would otherwise
	// leak forever on a never-cancelled queueCtx).
	c.mu.Lock()
	if c.shuttingDown {
		c.mu.Unlock()
		doneCh <- execResult{Err: fmt.Errorf("context shutting down")}
		return doneCh
	}
	c.startQueue()
	c.inflight.Add(1)
	c.mu.Unlock()

	select {
	case c.queue <- job:
	case <-c.queueCtx.Done():
		c.inflight.Add(-1)
		doneCh <- execResult{Err: fmt.Errorf("context shutting down")}
	}
	return doneCh
}

func (c *Session) processQueue() {
	for {
		select {
		case <-c.queueCtx.Done():
			return
		case job, ok := <-c.queue:
			if !ok {
				return
			}
			// Decrement via defer so a panic in runJob (or the doneCh send) still
			// releases the inflight reservation — a leaked reservation keeps
			// hasInflightWork() true forever and blocks idle GC of this session
			// ([B4]). Exactly one decrement per consumed job, matching the Enqueue
			// increment; the closure scopes the defer to this single iteration.
			func() {
				defer c.inflight.Add(-1)
				result, err := c.runJob(job)
				if job.doneCh != nil {
					job.doneCh <- execResult{cmd: result, Err: err}
				}
			}()
		}
	}
}

// IsBusy reports whether the context currently has an exec in flight. Used by the
// manager's /load aggregation.
func (c *Session) IsBusy() bool { return c.busy.Load() > 0 }

// hasInflightWork reports whether the context has any exec queued or running. Unlike
// IsBusy (which only reflects an exec actively executing), this also counts jobs
// accepted by Enqueue but not yet picked up by the queue goroutine — used by the idle
// sweeper so it never deletes a context with accepted-but-unstarted work.
func (c *Session) hasInflightWork() bool { return c.inflight.Load() > 0 }

// runJob is the per-execution sequence: tag the active command, send the worker
// the exec frame, await the terminal control chunk (or a timeout), then return.
func (c *Session) runJob(job execJob) (*CommandExecution, error) {
	c.busy.Add(1)
	defer c.busy.Add(-1)

	cmdID := uuid.NewString()
	exec := &CommandExecution{
		ID:        cmdID,
		Code:      job.code,
		Status:    CommandStatusRunning,
		StartedAt: time.Now(),
	}

	c.commandMu.Lock()
	c.activeCommand = exec
	c.commandMu.Unlock()

	c.touchLastUsed()

	cmd := WorkerCommand{
		Op:        "exec",
		SessionID: c.info.ID,
		ID:        cmdID,
		Code:      job.code,
		Envs:      job.envs,
		Reset:     job.reset,
	}
	if job.timeout > 0 {
		cmd.ExecTimeoutMS = job.timeout.Milliseconds()
	}

	if err := c.worker.Send(cmd); err != nil {
		exec.Status = CommandStatusError
		now := time.Now()
		exec.EndedAt = &now
		exec.Error = &Error{Name: "WorkerError", Value: err.Error()}
		c.commandMu.Lock()
		c.activeCommand = nil
		c.commandMu.Unlock()
		return exec, err
	}

	// Wait for the chunk handler to flip activeCommand off CommandStatusRunning,
	// or for the timeout to elapse and force-interrupt the worker.
	pollDone := make(chan struct{})
	go func() {
		ticker := time.NewTicker(20 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			c.commandMu.Lock()
			done := c.activeCommand == nil || c.activeCommand.Status != CommandStatusRunning
			c.commandMu.Unlock()
			if done {
				close(pollDone)
				return
			}
		}
	}()

	var timeoutC <-chan time.Time
	if job.timeout > 0 {
		t := time.NewTimer(job.timeout)
		defer t.Stop()
		timeoutC = t.C
	}

	select {
	case <-pollDone:
		c.commandMu.Lock()
		result := c.activeCommand
		c.activeCommand = nil
		c.commandMu.Unlock()
		c.touchLastUsed()
		return result, nil

	case <-timeoutC:
		_ = c.worker.Interrupt()

		grace := time.NewTimer(gracePeriod)
		defer grace.Stop()
		select {
		case <-pollDone:
			c.commandMu.Lock()
			result := c.activeCommand
			c.activeCommand = nil
			c.commandMu.Unlock()
			c.touchLastUsed()
			return result, nil
		case <-grace.C:
			c.commandMu.Lock()
			result := c.activeCommand
			if result != nil {
				now := time.Now()
				result.Status = CommandStatusTimeout
				result.EndedAt = &now
				result.Error = &Error{
					Name:  "TimeoutError",
					Value: "Execution timeout - code took too long to execute",
				}
			}
			c.activeCommand = nil
			c.commandMu.Unlock()
			c.touchLastUsed()
			return result, nil
		}
	}
}

// handleChunk is invoked by the Worker for every chunk it produces. It updates
// internal command state for terminal/error chunks and forwards every chunk to
// the attached WebSocket client (if any).
func (c *Session) handleChunk(chunk *WorkerChunk) {
	switch chunk.Type {
	case ChunkTypeError:
		c.commandMu.Lock()
		if c.activeCommand != nil {
			now := time.Now()
			c.activeCommand.Status = CommandStatusError
			c.activeCommand.EndedAt = &now
			c.activeCommand.Error = &Error{
				Name:      chunk.Name,
				Value:     chunk.Value,
				Traceback: chunk.Traceback,
			}
		}
		c.commandMu.Unlock()

	case ChunkTypeControl:
		c.commandMu.Lock()
		if c.activeCommand != nil {
			now := time.Now()
			switch chunk.Text {
			case ControlChunkTypeCompleted:
				if c.activeCommand.Status == CommandStatusRunning {
					c.activeCommand.Status = CommandStatusOK
				}
				c.activeCommand.EndedAt = &now
			case ControlChunkTypeInterrupted:
				c.activeCommand.Status = CommandStatusTimeout
				c.activeCommand.EndedAt = &now
			}
		}
		c.commandMu.Unlock()
	}

	c.emit(&OutputMessage{
		Type:      chunk.Type,
		Text:      chunk.Text,
		Name:      chunk.Name,
		Value:     chunk.Value,
		Traceback: chunk.Traceback,
		Formats:   chunk.Formats,
		Data:      chunk.Data,
	})
}

func (c *Session) touchLastUsed() {
	c.mu.Lock()
	c.info.LastUsedAt = time.Now()
	c.mu.Unlock()
}

// emit pushes a chunk to the currently attached WebSocket client (if any).
func (c *Session) emit(msg *OutputMessage) {
	c.mu.Lock()
	cl := c.client
	c.mu.Unlock()
	if cl == nil {
		return
	}
	select {
	case cl.send <- wsFrame{output: msg}:
	default:
		// Slow consumer — close it so the WS layer can recover.
		cl.requestClose(websocket.ClosePolicyViolation, "slow consumer")
		c.mu.Lock()
		if c.client != nil && c.client.id == cl.id {
			c.client = nil
		}
		c.mu.Unlock()
	}
}

// shutdown stops the queue, tears down the worker, and closes any attached client.
func (c *Session) shutdown() {
	c.mu.Lock()
	// Flip shuttingDown before cancelling the queue context so any Enqueue that
	// has not yet reserved its inflight slot is refused (see Enqueue). Combined
	// with the drain-until-inflight-zero loop in drainAndClose, this guarantees
	// every accepted job is answered and its reservation released.
	c.shuttingDown = true
	queue := c.queue
	stop := c.queueStop
	worker := c.worker
	cl := c.client
	c.client = nil
	c.info.Active = false
	c.mu.Unlock()

	if stop != nil {
		stop()
	}
	if queue != nil {
		// Drain pending jobs so their callers get an error rather than blocking forever.
		go c.drainAndClose(queue)
	}
	if worker != nil {
		// Best-effort teardown: log the error rather than propagate, since
		// shutdown() is invoked from delete/idle-sweep paths that have nothing
		// to do with a failed worker teardown beyond surfacing it.
		if err := worker.Shutdown(); err != nil && c.logger != nil {
			c.logger.Warn("worker shutdown error",
				slog.String("id", c.info.ID),
				slog.String("error", err.Error()),
			)
		}
	}
	if cl != nil {
		cl.requestClose(websocket.CloseGoingAway, "context shutdown")
	}
}

// drainAndClose answers every job accepted by Enqueue but not yet completed,
// returning a shutdown error to its caller and releasing its inflight
// reservation. It runs until inflight reaches 0 — a safe terminal because
// shutdown() set shuttingDown (under c.mu) before launching this, so no new
// Enqueue can reserve another slot, and the count therefore decreases
// monotonically to 0 and stays there. Jobs already picked up by processQueue are
// counted too (its deferred Add(-1) releases them), so the loop naturally waits
// for an in-flight runJob to finish once the worker teardown unblocks it.
//
// The channel is deliberately NOT closed: multiple concurrent Enqueue senders
// make a safe close impossible (a send on a closed channel panics), and leaving
// it open lets a send that wins Enqueue's select race land in the buffer where
// this loop still drains it.
func (c *Session) drainAndClose(q chan execJob) {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	// Bound the total wait: worker teardown (see shutdown) unblocks any in-flight
	// runJob, so inflight settles within ~gracePeriod. The deadline guarantees this
	// goroutine can't poll indefinitely should a worker somehow never complete.
	deadline := time.NewTimer(30 * time.Second)
	defer deadline.Stop()
	for {
		// Fast path: drain everything currently buffered.
		select {
		case job, ok := <-q:
			if !ok {
				return
			}
			if job.doneCh != nil {
				job.doneCh <- execResult{Err: fmt.Errorf("context shutdown")}
			}
			c.inflight.Add(-1)
			continue
		default:
		}
		// Buffer momentarily empty. With no reservations outstanding we are done.
		if c.inflight.Load() == 0 {
			return
		}
		// A reservation is still outstanding: either an in-flight runJob, or an
		// Enqueue that reserved before shuttingDown was set and has not resolved
		// its select yet. Wait for a job to land (drain it) or for the count to
		// fall (its queueCtx.Done branch / runJob completion), then re-check.
		select {
		case job, ok := <-q:
			if !ok {
				return
			}
			if job.doneCh != nil {
				job.doneCh <- execResult{Err: fmt.Errorf("context shutdown")}
			}
			c.inflight.Add(-1)
		case <-ticker.C:
		case <-deadline.C:
			if c.logger != nil {
				c.logger.Warn("drainAndClose giving up with jobs still in flight",
					slog.String("id", c.info.ID),
					slog.Int64("inflight", c.inflight.Load()))
			}
			return
		}
	}
}

// WSClient is the public alias of the internal wsClient handle returned by
// AttachWebSocket so the server package can call RequestClose on it.
type WSClient = wsClient

// AttachWebSocket connects a WebSocket client to the context. Only one client
// at a time — a new attach evicts the old one.
func (c *Session) AttachWebSocket(ws *websocket.Conn, logger logTarget) *WSClient {
	clientID := uuid.NewString()
	cl := &wsClient{
		id:     clientID,
		conn:   ws,
		send:   make(chan wsFrame, 1024),
		done:   make(chan struct{}),
		logger: logger,
	}

	c.mu.Lock()
	if c.client != nil {
		old := c.client
		c.client = cl
		c.mu.Unlock()
		old.requestClose(websocket.CloseGoingAway, "evicted by new client")
	} else {
		c.client = cl
		c.mu.Unlock()
	}

	go cl.writer()
	return cl
}

type logTarget interface {
	Debug(msg string, args ...any)
	Warn(msg string, args ...any)
}

// wsClient state lives in a dedicated file so the WS framing is in one place.
var _ = sync.Mutex{}
