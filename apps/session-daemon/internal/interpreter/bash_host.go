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

//go:embed repl_bash_host.js
var bashHostScript string

// bashCallTimeout bounds a single bridge bash() call end-to-end on the Go side.
// The host also applies its own per-exec AbortController timeout; this is the
// outer guard so a wedged host can't block a Python/TS bridge call forever.
const bashCallTimeout = 60 * time.Second

// BashFactory creates one shared Node "bash host" process per daemon and carves
// out per-session just-bash shells that multiplex over its stdin/stdout
// JSON-line protocol — the same shared-host strategy as TSFactory. Each session
// is one `new Bash({ fs: OverlayFs(/workspace) })`: reads hit the real
// filesystem, writes are private + in-memory, no real binaries or subprocesses.
//
// Besides serving standalone bash isolates via the Worker contract, the factory
// exposes Call/Release so the Python bash() bridge can route one-shot commands
// to the same host (the TS bridge runs just-bash in its own host, in-process).
type BashFactory struct {
	cfg    *config.Config
	logger *slog.Logger

	mu          sync.Mutex
	host        *bashHost
	hostScript  string
	replyMu     sync.Mutex
	pendingReps map[string]chan *WorkerChunk
	replyN      atomic.Uint64
}

func NewBashFactory(cfg *config.Config, logger *slog.Logger) (*BashFactory, error) {
	scriptPath, err := writeHostScript(cfg.HostScriptCacheDir, "daytona_session_repl_bash_host.js", bashHostScript)
	if err != nil {
		return nil, err
	}
	return &BashFactory{
		cfg:         cfg,
		logger:      logger.With(slog.String("component", "bash_factory")),
		hostScript:  scriptPath,
		pendingReps: make(map[string]chan *WorkerChunk),
	}, nil
}

func (f *BashFactory) Create(ctxID string, _ CreateSessionRequest, onChunk func(*WorkerChunk)) (Worker, error) {
	host, err := f.ensureHost()
	if err != nil {
		return nil, err
	}

	host.register(ctxID, onChunk)

	// Await the host's create acknowledgment so a host-side create failure
	// surfaces here instead of returning a live-but-broken worker (mirrors the
	// TS host). The host replies with a "created" control chunk or an error
	// chunk carrying the sessionId.
	ack := make(chan *WorkerChunk, 1)
	host.registerCreate(ctxID, ack)
	defer host.unregisterCreate(ctxID)

	if err := host.send(WorkerCommand{Op: "create", SessionID: ctxID}); err != nil {
		host.unregister(ctxID)
		return nil, err
	}

	t := time.NewTimer(createAckTimeout)
	defer t.Stop()
	select {
	case chunk := <-ack:
		if chunk.Type == ChunkTypeError {
			host.unregister(ctxID)
			return nil, fmt.Errorf("bash host create failed: %s: %s", chunk.Name, chunk.Value)
		}
	case <-host.done:
		host.unregister(ctxID)
		return nil, errors.New("bash host: exited before create acknowledgment")
	case <-t.C:
		host.unregister(ctxID)
		return nil, errors.New("bash host: create acknowledgment timeout")
	}

	w := &bashHostWorker{ctxID: ctxID, host: host}
	w.active.Store(true)
	return w, nil
}

// ListPackages has no real meaning for the virtual bash shell (there is no
// installable package set), so we return an empty catalog rather than erroring —
// the /packages endpoint stays well-formed for language=bash.
func (f *BashFactory) ListPackages() ([]PackageInfo, error) {
	return []PackageInfo{}, nil
}

func (f *BashFactory) Shutdown() {
	defer removeHostScriptDir(f.hostScript)
	f.mu.Lock()
	host := f.host
	f.host = nil
	f.mu.Unlock()
	if host != nil {
		host.shutdown()
	}
}

// Call runs a single bash command to completion on the shared host and returns
// the aggregated result. It is the routing primitive behind the Python bash()
// bridge. The host keeps a per-session shell (sessionID) so overlay writes
// persist across calls within a session; Release drops it.
func (f *BashFactory) Call(sessionID, code string, env map[string]string) (stdout, stderr string, exitCode int, err error) {
	host, herr := f.ensureHost()
	if herr != nil {
		return "", "", 0, herr
	}

	replyID := fmt.Sprintf("bash-call-%d", f.replyN.Add(1))
	ch := make(chan *WorkerChunk, 1)
	f.replyMu.Lock()
	f.pendingReps[replyID] = ch
	f.replyMu.Unlock()
	defer func() {
		f.replyMu.Lock()
		delete(f.pendingReps, replyID)
		f.replyMu.Unlock()
	}()

	if serr := host.send(WorkerCommand{Op: "bash-call", SessionID: sessionID, Code: code, Envs: env, Reply: replyID}); serr != nil {
		return "", "", 0, serr
	}

	t := time.NewTimer(bashCallTimeout)
	defer t.Stop()
	select {
	case chunk := <-ch:
		if chunk.Text == "bash-call-error" {
			return "", "", 0, errors.New(chunk.Value)
		}
		return chunk.Stdout, chunk.Stderr, chunk.ExitCode, nil
	case <-host.done:
		return "", "", 0, errors.New("bash host: exited before call result")
	case <-t.C:
		return "", "", 0, errors.New("bash host: call timeout")
	}
}

// Release drops the per-session shell backing a bridge caller (e.g. when a
// Python session is torn down). Best-effort: a missing host/session is a no-op.
func (f *BashFactory) Release(sessionID string) {
	f.mu.Lock()
	host := f.host
	f.mu.Unlock()
	if host == nil || !host.active.Load() {
		return
	}
	_ = host.send(WorkerCommand{Op: "delete", SessionID: sessionID})
}

func (f *BashFactory) ensureHost() (*bashHost, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.host != nil && f.host.active.Load() {
		return f.host, nil
	}

	parentCtx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(parentCtx, f.cfg.NodeInterpreter, f.hostScript)
	hostNodeModules := filepath.Join(f.cfg.NodeBundleRoot, "node_modules")
	cmd.Env = append(os.Environ(),
		"SESSION_DAEMON_WORKSPACE_ROOT="+f.cfg.WorkspaceRoot,
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
		return nil, fmt.Errorf("start bash host: %w", err)
	}

	host := &bashHost{
		factory:        f,
		cmd:            cmd,
		stdin:          stdin,
		stdout:         stdout,
		cancel:         cancel,
		done:           make(chan struct{}),
		logger:         f.logger.With(slog.String("component", "bash_host"), slog.Int("pid", cmd.Process.Pid)),
		listeners:      make(map[string]func(*WorkerChunk)),
		pendingCreates: make(map[string]chan *WorkerChunk),
	}
	host.active.Store(true)

	go host.readLoop()
	go host.waitLoop()
	f.host = host
	return host, nil
}

// bashHost is the long-lived Node process running repl_bash_host.js. It demuxes
// streaming chunks by sessionId and routes bash-call replies + create acks to
// their waiters.
type bashHost struct {
	factory *BashFactory
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	cancel  context.CancelFunc
	done    chan struct{}
	logger  *slog.Logger

	writeMu sync.Mutex

	mu             sync.Mutex
	listeners      map[string]func(*WorkerChunk)
	pendingCreates map[string]chan *WorkerChunk
	active         activeFlag
}

func (h *bashHost) register(ctxID string, onChunk func(*WorkerChunk)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.listeners[ctxID] = onChunk
}

func (h *bashHost) unregister(ctxID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.listeners, ctxID)
}

func (h *bashHost) registerCreate(ctxID string, ch chan *WorkerChunk) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.pendingCreates[ctxID] = ch
}

func (h *bashHost) unregisterCreate(ctxID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.pendingCreates, ctxID)
}

func (h *bashHost) deliverCreate(chunk *WorkerChunk) bool {
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

func (h *bashHost) send(cmd WorkerCommand) error {
	h.writeMu.Lock()
	defer h.writeMu.Unlock()
	if !h.active.Load() {
		return errors.New("bash host: not active")
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

func (h *bashHost) shutdown() {
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

func (h *bashHost) readLoop() {
	// Bounded reader (not bufio.Scanner) so an over-long line on this SHARED host
	// is drained + skipped rather than ending the loop and taking down every bash
	// session — same rationale as ts_host.go.
	reader := bufio.NewReaderSize(h.stdout, 64*1024)
	for {
		line, err := readBoundedLine(reader)
		if errors.Is(err, errLineTooLong) {
			h.logger.Warn("dropping oversized bash host chunk (no listener notified)", slog.Int("limit", maxWorkerLineBytes))
			continue
		}
		if len(line) == 0 {
			if err != nil {
				if !errors.Is(err, io.EOF) {
					h.logger.Debug("bash host readLoop ended", slog.String("error", err.Error()))
				}
				return
			}
			continue
		}
		var chunk WorkerChunk
		if jerr := json.Unmarshal([]byte(line), &chunk); jerr != nil {
			h.logger.Warn("malformed bash host chunk", slog.String("error", jerr.Error()))
		} else {
			h.dispatchChunk(&chunk)
		}
		if err != nil {
			if !errors.Is(err, io.EOF) {
				h.logger.Debug("bash host readLoop ended", slog.String("error", err.Error()))
			}
			return
		}
	}
}

func (h *bashHost) dispatchChunk(chunk *WorkerChunk) {
	// bash-call replies route to the factory's pending-request table.
	if chunk.Type == ChunkTypeControl && (chunk.Text == "bash-call-result" || chunk.Text == "bash-call-error") {
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
	// Lifecycle control chunks ("created"/"deleted"/"host-ready") are not
	// user-visible; only "completed"/"interrupted" reach the per-session handler.
	if chunk.Type == ChunkTypeControl && chunk.Text != ControlChunkTypeCompleted &&
		chunk.Text != ControlChunkTypeInterrupted {
		return
	}
	if chunk.SessionID == "" {
		h.logger.Debug("dropping bash chunk without sessionId")
		return
	}
	h.mu.Lock()
	listener := h.listeners[chunk.SessionID]
	h.mu.Unlock()
	if listener != nil {
		listener(chunk)
	}
}

func (h *bashHost) waitLoop() {
	err := h.cmd.Wait()
	h.active.Store(false)
	close(h.done)
	if err != nil {
		h.logger.Warn("bash host exited", slog.String("error", err.Error()))
	}

	// The host is gone, so any in-flight exec waiting on a "completed" control
	// chunk would block forever. Synthesize a WorkerProcessError + completed pair
	// for every registered listener so the waiter in session.go unblocks
	// (mirrors the TS host).
	msg := "bash host exited"
	if err != nil {
		msg = "bash host exited: " + err.Error()
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
		fn(&WorkerChunk{SessionID: id, Type: ChunkTypeError, Name: "WorkerProcessError", Value: msg})
		fn(&WorkerChunk{SessionID: id, Type: ChunkTypeControl, Text: ControlChunkTypeCompleted})
	}
}

// bashHostWorker is the per-session Worker view of the shared bash host.
type bashHostWorker struct {
	ctxID  string
	host   *bashHost
	active activeFlag
}

func (w *bashHostWorker) Active() bool { return w.active.Load() }

func (w *bashHostWorker) Send(cmd WorkerCommand) error {
	if !w.active.Load() {
		return errors.New("worker closed")
	}
	cmd.SessionID = w.ctxID
	return w.host.send(cmd)
}

func (w *bashHostWorker) Interrupt() error {
	return w.host.send(WorkerCommand{Op: "interrupt", SessionID: w.ctxID})
}

func (w *bashHostWorker) Shutdown() error {
	if !w.active.Swap(false) {
		return nil
	}
	err := w.host.send(WorkerCommand{Op: "delete", SessionID: w.ctxID})
	w.host.unregister(w.ctxID)
	return err
}
