// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package interpreter

import (
	"bufio"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/daytonaio/session-daemon/internal/config"
)

//go:embed repl_host.js
var typescriptHostScript string

// TSFactory creates one shared Node host process per daemon, then carves out
// per-context "workers" that multiplex over the host's stdin/stdout JSON-line
// protocol. This is the V8-session strategy described in plan §3 — many
// contexts in one host means ~10MB per context vs ~50MB for subprocess workers.
type TSFactory struct {
	cfg    *config.Config
	logger *slog.Logger

	mu          sync.Mutex
	host        *tsHost
	hostScript  string
	pkgsCache   []PackageInfo
	pkgsAt      time.Time
	replyMu     sync.Mutex
	pendingReps map[string]chan *WorkerChunk
	replyN      atomic.Uint64
}

func NewTSFactory(cfg *config.Config, logger *slog.Logger) (*TSFactory, error) {
	scriptPath, err := writeHostScript(cfg.HostScriptCacheDir, "daytona_session_repl_host.js", typescriptHostScript)
	if err != nil {
		return nil, err
	}
	return &TSFactory{
		cfg:         cfg,
		logger:      logger.With(slog.String("component", "ts_factory")),
		hostScript:  scriptPath,
		pendingReps: make(map[string]chan *WorkerChunk),
	}, nil
}

func (f *TSFactory) Create(ctxID string, req CreateSessionRequest, onChunk func(*WorkerChunk)) (Worker, error) {
	host, err := f.ensureHost()
	if err != nil {
		return nil, err
	}

	memMB := req.MemoryLimitMB
	if memMB <= 0 {
		memMB = f.cfg.TSDefaultMemoryMB
	}
	if memMB > f.cfg.TSMaxMemoryMB {
		return nil, fmt.Errorf("memoryLimitMb %d exceeds cap %d", memMB, f.cfg.TSMaxMemoryMB)
	}

	host.register(ctxID, onChunk)

	// Await the host's create acknowledgment so a host-side create failure
	// (ContextExistsError/HostError from repl_host.js) surfaces as an error
	// here instead of returning a live-but-broken worker. The host replies with
	// either a "created" control chunk or an error chunk carrying the sessionId.
	ack := make(chan *WorkerChunk, 1)
	host.registerCreate(ctxID, ack)
	defer host.unregisterCreate(ctxID)

	if err := host.send(WorkerCommand{Op: "create", SessionID: ctxID, MemoryLimitMB: memMB}); err != nil {
		host.unregister(ctxID)
		return nil, err
	}

	t := time.NewTimer(createAckTimeout)
	defer t.Stop()
	select {
	case chunk := <-ack:
		if chunk.Type == ChunkTypeError {
			host.unregister(ctxID)
			return nil, fmt.Errorf("ts host create failed: %s: %s", chunk.Name, chunk.Value)
		}
	case <-host.done:
		host.unregister(ctxID)
		return nil, errors.New("ts host: exited before create acknowledgment")
	case <-t.C:
		host.unregister(ctxID)
		return nil, errors.New("ts host: create acknowledgment timeout")
	}

	w := &tsHostWorker{ctxID: ctxID, host: host}
	w.active.Store(true)
	return w, nil
}

func (f *TSFactory) ListPackages() ([]PackageInfo, error) {
	host, err := f.ensureHost()
	if err != nil {
		return nil, err
	}
	f.mu.Lock()
	if time.Since(f.pkgsAt) < 5*time.Minute && f.pkgsCache != nil {
		out := f.pkgsCache
		f.mu.Unlock()
		return out, nil
	}
	f.mu.Unlock()

	replyID := fmt.Sprintf("list-packages-%d", f.replyN.Add(1))
	ch := make(chan *WorkerChunk, 1)
	f.replyMu.Lock()
	f.pendingReps[replyID] = ch
	f.replyMu.Unlock()
	defer func() {
		f.replyMu.Lock()
		delete(f.pendingReps, replyID)
		f.replyMu.Unlock()
	}()

	if err := host.send(WorkerCommand{Op: "list-packages", Reply: replyID}); err != nil {
		return nil, err
	}

	select {
	case chunk := <-ch:
		f.mu.Lock()
		f.pkgsCache = chunk.Packages
		f.pkgsAt = time.Now()
		f.mu.Unlock()
		return chunk.Packages, nil
	case <-time.After(15 * time.Second):
		return nil, errors.New("ts host: list-packages timeout")
	}
}

func (f *TSFactory) Shutdown() {
	defer removeHostScriptDir(f.hostScript)
	f.mu.Lock()
	host := f.host
	f.host = nil
	f.mu.Unlock()
	if host != nil {
		host.shutdown()
	}
}

func (f *TSFactory) ensureHost() (*tsHost, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.host != nil && f.host.active.Load() {
		return f.host, nil
	}

	parentCtx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(parentCtx, f.cfg.NodeInterpreter, f.hostScript)
	// NODE_PATH points Node's module resolver at the host-side node_modules baked
	// into the image (isolated-vm + esbuild-wasm). The script itself is written
	// to a temp dir for portability, so we must explicitly tell Node where its
	// dependencies live — adjacent-directory resolution from /tmp does not find
	// /usr/lib/daytona/repl_host/node_modules.
	hostNodeModules := filepath.Join(f.cfg.NodeBundleRoot, "node_modules")
	cmd.Env = append(os.Environ(),
		"SESSION_DAEMON_USER_NODE_MODULES_ROOT="+f.cfg.WorkspaceRoot,
		"NODE_PATH="+hostNodeModules,
	)
	cmd.Dir = f.cfg.NodeBundleRoot

	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		cancel()
		return nil, err
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		_ = stdout.Close()
		cancel()
		return nil, fmt.Errorf("start ts host: %w", err)
	}

	host := &tsHost{
		factory:        f,
		cmd:            cmd,
		stdin:          stdin,
		stdout:         stdout,
		cancel:         cancel,
		done:           make(chan struct{}),
		logger:         f.logger.With(slog.String("component", "ts_host"), slog.Int("pid", cmd.Process.Pid)),
		listeners:      make(map[string]func(*WorkerChunk)),
		pendingCreates: make(map[string]chan *WorkerChunk),
	}
	host.active.Store(true)

	go host.readLoop()
	go host.waitLoop()
	f.host = host
	return host, nil
}

// tsHost is the long-lived Node process. It demuxes incoming chunks by
// `sessionId` and dispatches them to the chunk handler the corresponding
// worker registered at create time.
type tsHost struct {
	factory *TSFactory
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	cancel  context.CancelFunc
	done    chan struct{}
	logger  *slog.Logger

	writeMu sync.Mutex

	mu        sync.Mutex
	listeners map[string]func(*WorkerChunk)
	// pendingCreates routes the host's create acknowledgment (a "created" control
	// chunk or a ContextExistsError/HostError) back to the in-flight Create call,
	// keyed by sessionId. Guarded by mu.
	pendingCreates map[string]chan *WorkerChunk
	active         activeFlag
}

func (h *tsHost) register(ctxID string, onChunk func(*WorkerChunk)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.listeners[ctxID] = onChunk
}

func (h *tsHost) unregister(ctxID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.listeners, ctxID)
}

func (h *tsHost) registerCreate(ctxID string, ch chan *WorkerChunk) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.pendingCreates[ctxID] = ch
}

func (h *tsHost) unregisterCreate(ctxID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.pendingCreates, ctxID)
}

// deliverCreate routes a create acknowledgment chunk to a waiting Create call.
// Returns true when the chunk was consumed as an ack (and must not be dispatched
// to the per-context listener).
func (h *tsHost) deliverCreate(chunk *WorkerChunk) bool {
	h.mu.Lock()
	ch := h.pendingCreates[chunk.SessionID]
	h.mu.Unlock()
	if ch == nil {
		return false
	}
	select {
	case ch <- chunk:
	default:
	}
	return true
}

func (h *tsHost) send(cmd WorkerCommand) error {
	h.writeMu.Lock()
	defer h.writeMu.Unlock()
	if !h.active.Load() {
		return errors.New("ts host: not active")
	}
	data, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if _, err := h.stdin.Write(data); err != nil {
		return err
	}
	return nil
}

func (h *tsHost) shutdown() {
	if !h.active.Swap(false) {
		return
	}
	if h.stdin != nil {
		_ = h.stdin.Close()
	}
	if h.cmd != nil && h.cmd.Process != nil {
		_ = h.cmd.Process.Signal(os.Interrupt)
	}
	t := time.NewTimer(gracePeriod)
	defer t.Stop()
	select {
	case <-h.done:
	case <-t.C:
		if h.cmd != nil && h.cmd.Process != nil {
			_ = h.cmd.Process.Kill()
		}
	}
	if h.cancel != nil {
		h.cancel()
	}
}

// Swap behavior for activeFlag (compat helper).
func (a *activeFlag) Swap(v bool) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	prev := a.v
	a.v = v
	return prev
}

func (h *tsHost) readLoop() {
	// Bounded reader rather than bufio.Scanner: a Scanner returns ErrTooLong on a
	// line past its cap and ENDS the loop, which — via waitLoop's host-exit
	// propagation — would take down EVERY context sharing this host ([C3]). With
	// readBoundedLine an over-long line is drained and SKIPPED (no listener is
	// notified, see reportOversizedLine), leaving the shared host and all other
	// contexts intact. The emit-side frame guard (per-field MAX_CHUNK_BYTES plus a
	// whole-line MAX_LINE_BYTES cap in repl_host.js) makes this unreachable in
	// practice, so the affected exec is protected without fanning a failure out.
	reader := bufio.NewReaderSize(h.stdout, 64*1024)
	for {
		line, err := readBoundedLine(reader)
		if errors.Is(err, errLineTooLong) {
			h.reportOversizedLine()
			continue
		}
		if len(line) == 0 {
			if err != nil {
				if !errors.Is(err, io.EOF) {
					h.logger.Debug("ts host readLoop ended", slog.String("error", err.Error()))
				}
				return
			}
			continue
		}
		var chunk WorkerChunk
		if jerr := json.Unmarshal([]byte(line), &chunk); jerr != nil {
			h.logger.Warn("malformed ts host chunk", slog.String("error", jerr.Error()))
		} else {
			h.dispatchChunk(&chunk)
		}
		if err != nil {
			if !errors.Is(err, io.EOF) {
				h.logger.Debug("ts host readLoop ended", slog.String("error", err.Error()))
			}
			return
		}
	}
}

// dispatchChunk routes a single decoded chunk to the reply table, an in-flight
// Create, or the per-context listener, filtering non-user-visible lifecycle
// control chunks. Split out of readLoop so the read loop can stay focused on
// framing/recovery.
func (h *tsHost) dispatchChunk(chunk *WorkerChunk) {
	// Reply chunks (e.g., list-packages) are routed to the factory's reply table.
	if chunk.Type == ChunkTypeControl && chunk.Text == "list-packages-result" {
		h.factory.replyMu.Lock()
		ch := h.factory.pendingReps[chunk.Reply]
		h.factory.replyMu.Unlock()
		if ch != nil {
			ch <- chunk
		}
		return
	}
	// Create acknowledgment ("created" control or a create-time error) is routed
	// back to the in-flight Create call rather than to the per-context listener.
	if chunk.SessionID != "" &&
		((chunk.Type == ChunkTypeControl && chunk.Text == "created") || chunk.Type == ChunkTypeError) {
		if h.deliverCreate(chunk) {
			return
		}
	}
	// Lifecycle control chunks ("created"/"deleted"/"interrupted"/"host-ready") are
	// not user-visible; only "completed" must reach the per-context handler.
	if chunk.Type == ChunkTypeControl && chunk.Text != ControlChunkTypeCompleted &&
		chunk.Text != ControlChunkTypeInterrupted {
		return
	}
	// Per-context chunk → look up the listener and dispatch.
	if chunk.SessionID == "" {
		h.logger.Debug("dropping chunk without sessionId")
		return
	}
	h.mu.Lock()
	listener := h.listeners[chunk.SessionID]
	h.mu.Unlock()
	if listener != nil {
		listener(chunk)
	}
}

// reportOversizedLine handles an over-long output frame on the SHARED host by
// draining and skipping it with a logged warning, WITHOUT notifying any listener.
// The sessionId lived inside the unparseable (already-drained) frame, so we
// cannot know which context emitted it — fanning a hard failure out to every
// registered context would take down unrelated in-flight execs on this shared
// host ([C3] cross-session DoS). Dropping the line silently keeps the host and
// all other contexts alive. The affected (unknown) exec is protected on the EMIT
// side instead: repl_host.js bounds every serialized frame below the reader cap
// (per-field MAX_CHUNK_BYTES plus a whole-line MAX_LINE_BYTES guard), so this
// path is effectively unreachable for well-behaved output and no exec hangs.
func (h *tsHost) reportOversizedLine() {
	h.logger.Warn("dropping oversized ts host chunk (no listener notified)", slog.Int("limit", maxWorkerLineBytes))
}

func (h *tsHost) waitLoop() {
	err := h.cmd.Wait()
	h.active.Store(false)
	close(h.done)
	if err != nil {
		h.logger.Warn("ts host exited", slog.String("error", err.Error()))
	}

	// The host is gone, so any in-flight exec waiting on a "completed" control
	// chunk would block forever. Mirror the python worker's waitLoop: synthesize
	// a WorkerProcessError + completed pair for every registered listener so the
	// waiter in session.go unblocks. Snapshot under the lock to avoid racing with
	// (un)register, then dispatch without holding it.
	msg := "ts host exited"
	if err != nil {
		msg = "ts host exited: " + err.Error()
	}
	h.mu.Lock()
	listeners := make(map[string]func(*WorkerChunk), len(h.listeners))
	for id, fn := range h.listeners {
		listeners[id] = fn
	}
	h.mu.Unlock()
	for id, fn := range listeners {
		if fn == nil {
			continue
		}
		fn(&WorkerChunk{
			SessionID: id,
			Type:      ChunkTypeError,
			Name:      "WorkerProcessError",
			Value:     msg,
		})
		fn(&WorkerChunk{
			SessionID: id,
			Type:      ChunkTypeControl,
			Text:      ControlChunkTypeCompleted,
		})
	}
}

// tsHostWorker is the per-context Worker view of the shared host.
type tsHostWorker struct {
	ctxID  string
	host   *tsHost
	active activeFlag
}

func (w *tsHostWorker) Active() bool { return w.active.Load() }

func (w *tsHostWorker) Send(cmd WorkerCommand) error {
	if !w.active.Load() {
		return errors.New("worker closed")
	}
	cmd.SessionID = w.ctxID
	return w.host.send(cmd)
}

func (w *tsHostWorker) Interrupt() error {
	return w.host.send(WorkerCommand{Op: "interrupt", SessionID: w.ctxID})
}

func (w *tsHostWorker) Shutdown() error {
	if !w.active.Swap(false) {
		return nil
	}
	err := w.host.send(WorkerCommand{Op: "delete", SessionID: w.ctxID})
	w.host.unregister(w.ctxID)
	return err
}
