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
	"sync"
	"syscall"
	"time"

	"github.com/daytonaio/session-daemon/internal/config"
)

//go:embed repl_worker.py
var pythonWorkerScript string

// PythonFactory creates one CPython subprocess per context. Identical strategy
// to the existing daytona-daemon — chosen for v1 because subprocess isolation
// gives us full OS-level boundaries with zero engine-specific risk.
//
// To keep one-shot exec latency from being dominated by CPython startup
// (~300-400 ms incl. matplotlib import on most images), the factory also keeps
// a small bounded pool of pre-spawned, idle worker processes ready to be
// claimed by the next Create() call. See `warm` and `pool` below.
type PythonFactory struct {
	cfg        *config.Config
	logger     *slog.Logger
	workerPath string
	pkgsOnce   sync.Once
	pkgs       []PackageInfo
	pkgsErr    error

	// Pool of pre-spawned, idle workers. nil when PyWarmPoolSize <= 0.
	pool        chan *warmPython
	poolStop    chan struct{}
	poolWG      sync.WaitGroup
	poolStarted bool

	// bashInvoker backs the Python bash() bridge. nil when the bash engine is
	// unavailable, in which case bash() raises "bash runtime unavailable".
	bashInvoker BashInvoker
}

// SetBashInvoker wires the bash() bridge for Python workers. Called once at
// server construction (before any context is created), so no locking is needed.
func (f *PythonFactory) SetBashInvoker(inv BashInvoker) { f.bashInvoker = inv }

// warmPython is an already-spawned CPython REPL waiting to be wired up to a
// context. Its stdout has NOT been read yet (the REPL emits nothing until it
// receives its first command on stdin), so it's safe for the wrapper to start
// its own readLoop on handout.
type warmPython struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	cancel context.CancelFunc
}

func NewPythonFactory(cfg *config.Config, logger *slog.Logger) (*PythonFactory, error) {
	path, err := writeHostScript(cfg.WorkerScriptDir, "daytona_session_repl_worker.py", pythonWorkerScript)
	if err != nil {
		return nil, err
	}
	f := &PythonFactory{
		cfg:        cfg,
		logger:     logger.With(slog.String("component", "py_factory")),
		workerPath: path,
	}
	if cfg.PyWarmPoolSize > 0 {
		f.pool = make(chan *warmPython, cfg.PyWarmPoolSize)
		f.poolStop = make(chan struct{})
		f.poolStarted = true
		// Refill the pool to target. We fill once up-front instead of lazily
		// because the user-visible win is on the FIRST cold call — if we wait
		// for that first call to trigger a fill, the first call is slow.
		for i := 0; i < cfg.PyWarmPoolSize; i++ {
			f.scheduleRefill()
		}
	}
	return f, nil
}

func (f *PythonFactory) Create(ctxID string, _ CreateSessionRequest, onChunk func(*WorkerChunk)) (Worker, error) {
	warm, err := f.claimWarm()
	if err != nil {
		return nil, err
	}
	defer f.scheduleRefill()

	w := &pythonSubprocessWorker{
		ctxID:       ctxID,
		cmd:         warm.cmd,
		stdin:       warm.stdin,
		stdout:      warm.stdout,
		onChunk:     onChunk,
		cancel:      warm.cancel,
		done:        make(chan struct{}),
		bashInvoker: f.bashInvoker,
		logger:      f.logger.With(slog.String("ctx", ctxID), slog.Int("pid", warm.cmd.Process.Pid)),
	}
	w.active.Store(true)

	go w.readLoop()
	go w.waitLoop()

	return w, nil
}

// claimWarm pops a pre-spawned worker from the pool (non-blocking) or spawns
// one on the request path if the pool is empty/disabled. The fallback path
// keeps the daemon correct under bursty load without holding a request while
// we refill.
func (f *PythonFactory) claimWarm() (*warmPython, error) {
	if f.pool != nil {
		select {
		case warm := <-f.pool:
			return warm, nil
		default:
		}
	}
	return f.spawnWarm()
}

func (f *PythonFactory) spawnWarm() (*warmPython, error) {
	parentCtx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(parentCtx, f.cfg.PythonInterpreter, f.workerPath)
	cmd.Env = os.Environ()

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
		return nil, fmt.Errorf("start python worker: %w", err)
	}
	return &warmPython{cmd: cmd, stdin: stdin, stdout: stdout, cancel: cancel}, nil
}

// scheduleRefill kicks off a background spawn (and shoves the result into the
// pool channel) without blocking the caller. Drops the spawn if the pool is
// closed or already full.
func (f *PythonFactory) scheduleRefill() {
	if f.pool == nil {
		return
	}
	f.poolWG.Add(1)
	go func() {
		defer f.poolWG.Done()
		select {
		case <-f.poolStop:
			return
		default:
		}
		warm, err := f.spawnWarm()
		if err != nil {
			f.logger.Warn("warm python spawn failed", slog.String("error", err.Error()))
			return
		}
		select {
		case f.pool <- warm:
		case <-f.poolStop:
			// Factory shutting down: kill the warm process we just spawned.
			f.killWarm(warm)
		default:
			// Channel full (rare; means another refill landed first). Discard
			// to keep ourselves within the configured budget.
			f.killWarm(warm)
		}
	}()
}

func (f *PythonFactory) killWarm(warm *warmPython) {
	if warm == nil {
		return
	}
	if warm.stdin != nil {
		_ = warm.stdin.Close()
	}
	if warm.cmd != nil && warm.cmd.Process != nil {
		_ = warm.cmd.Process.Kill()
		_, _ = warm.cmd.Process.Wait()
	}
	if warm.cancel != nil {
		warm.cancel()
	}
}

func (f *PythonFactory) ListPackages() ([]PackageInfo, error) {
	f.pkgsOnce.Do(func() {
		out, err := exec.Command(f.cfg.PythonInterpreter, "-m", "pip", "list", "--format=json", "--disable-pip-version-check").Output()
		if err != nil {
			f.pkgsErr = err
			return
		}
		var raw []struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		}
		if err := json.Unmarshal(out, &raw); err != nil {
			f.pkgsErr = err
			return
		}
		pkgs := make([]PackageInfo, 0, len(raw))
		for _, r := range raw {
			pkgs = append(pkgs, PackageInfo{Name: r.Name, Version: r.Version})
		}
		f.pkgs = pkgs
	})
	return f.pkgs, f.pkgsErr
}

func (f *PythonFactory) Shutdown() {
	defer removeHostScriptDir(f.workerPath)
	// Active workers self-clean on Worker.Shutdown(). The factory still owns
	// any warm processes idling in the pool — close the stop channel so
	// pending refills bail, then drain + kill what's already in the channel.
	if !f.poolStarted {
		return
	}
	close(f.poolStop)
	f.poolWG.Wait()
	if f.pool == nil {
		return
	}
	close(f.pool)
	for warm := range f.pool {
		f.killWarm(warm)
	}
}

// pythonSubprocessWorker — one process per context.
type pythonSubprocessWorker struct {
	ctxID   string
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	onChunk func(*WorkerChunk)

	cancel context.CancelFunc
	done   chan struct{}

	mu      sync.Mutex
	stdinMu sync.Mutex // serializes writes to the worker's stdin (exec cmds + hostcall replies)
	closed  bool
	active  activeFlag

	bashInvoker BashInvoker
	logger      *slog.Logger
}

// hostCall is a request the Python worker emits on stdout when user code calls
// bash() — the bridge round-trip. The daemon routes it to the bash host and
// writes a hostCallResult back to the worker's stdin (correlated by ID).
type hostCall struct {
	ID     string            `json:"id"`
	Method string            `json:"method"`
	Cmd    string            `json:"cmd"`
	Env    map[string]string `json:"env"`
}

// hostCallResult is the reply written back to the worker's stdin. Type is
// "hostcall_result" on success or "hostcall_error" on failure.
type hostCallResult struct {
	Type     string `json:"type"`
	ID       string `json:"id"`
	Stdout   string `json:"stdout,omitempty"`
	Stderr   string `json:"stderr,omitempty"`
	ExitCode int    `json:"exitCode"`
	Message  string `json:"message,omitempty"`
}

// activeFlag is a tiny wrapper to avoid pulling sync/atomic.Bool in older toolchains.
type activeFlag struct {
	mu sync.Mutex
	v  bool
}

func (a *activeFlag) Store(b bool) { a.mu.Lock(); a.v = b; a.mu.Unlock() }
func (a *activeFlag) Load() bool   { a.mu.Lock(); defer a.mu.Unlock(); return a.v }

func (w *pythonSubprocessWorker) Active() bool { return w.active.Load() }

func (w *pythonSubprocessWorker) Send(cmd WorkerCommand) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return errors.New("worker closed")
	}

	// Translate WorkerCommand into the Python script's expected shape.
	if cmd.Op != "exec" && cmd.Op != "" {
		return fmt.Errorf("python worker: unsupported op %q", cmd.Op)
	}
	payload := map[string]interface{}{
		"id":   cmd.ID,
		"code": cmd.Code,
	}
	if cmd.Envs != nil {
		payload["envs"] = cmd.Envs
	}
	if cmd.Reset {
		payload["reset"] = true
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := w.writeStdin(data); err != nil {
		return fmt.Errorf("worker stdin: %w", err)
	}
	return nil
}

// writeStdin serializes all writes to the worker's stdin. Both exec commands
// (Send) and bash() hostcall replies (handleHostCall) go through here so they
// never interleave bytes on the pipe.
func (w *pythonSubprocessWorker) writeStdin(data []byte) error {
	w.stdinMu.Lock()
	defer w.stdinMu.Unlock()
	_, err := w.stdin.Write(data)
	return err
}

// handleHostCall services a bash() bridge request: run the command on the shared
// bash host and write the correlated result back to the worker's stdin. Runs in
// its own goroutine so a slow command never stalls the worker's read loop (and
// thus can't delay processing an interrupt). The Python worker is blocked on
// stdin awaiting this reply; if it was interrupted first, it discards the stray
// reply (see repl_worker.py handle_command).
func (w *pythonSubprocessWorker) handleHostCall(hc *hostCall) {
	res := hostCallResult{ID: hc.ID, Type: "hostcall_result"}
	switch {
	case hc.Method != "bash":
		res.Type = "hostcall_error"
		res.Message = "unknown hostcall method: " + hc.Method
	case w.bashInvoker == nil:
		res.Type = "hostcall_error"
		res.Message = "bash runtime unavailable"
	default:
		stdout, stderr, code, err := w.bashInvoker.Call(w.ctxID, hc.Cmd, hc.Env)
		if err != nil {
			res.Type = "hostcall_error"
			res.Message = err.Error()
		} else {
			res.Stdout, res.Stderr, res.ExitCode = stdout, stderr, code
		}
	}
	data, err := json.Marshal(res)
	if err != nil {
		return
	}
	data = append(data, '\n')
	if err := w.writeStdin(data); err != nil {
		w.logger.Debug("hostcall reply write failed", slog.String("error", err.Error()))
	}
}

func (w *pythonSubprocessWorker) Interrupt() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.cmd == nil || w.cmd.Process == nil {
		return nil
	}
	return w.cmd.Process.Signal(syscall.SIGINT)
}

func (w *pythonSubprocessWorker) Shutdown() error {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return nil
	}
	w.closed = true
	w.active.Store(false)
	cmd := w.cmd
	stdin := w.stdin
	cancel := w.cancel
	w.mu.Unlock()

	// Drop the per-session shell backing this context's bash() bridge (if any).
	if w.bashInvoker != nil {
		w.bashInvoker.Release(w.ctxID)
	}

	// Teardown is best-effort: the process/pipe may already be gone, which surfaces
	// as benign races (os.ErrProcessDone, os.ErrClosed). Only collect genuine errors
	// so callers don't log noise on every normal shutdown.
	var errs []error
	if stdin != nil {
		if err := stdin.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
			errs = append(errs, fmt.Errorf("close stdin: %w", err))
		}
	}
	if cmd != nil && cmd.Process != nil {
		if err := cmd.Process.Signal(syscall.SIGTERM); err != nil && !errors.Is(err, os.ErrProcessDone) {
			errs = append(errs, fmt.Errorf("signal SIGTERM: %w", err))
		}
	}
	t := time.NewTimer(gracePeriod)
	defer t.Stop()
	select {
	case <-w.done:
	case <-t.C:
		if cmd != nil && cmd.Process != nil {
			if err := cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
				errs = append(errs, fmt.Errorf("kill process: %w", err))
			}
		}
	}
	if cancel != nil {
		cancel()
	}
	return errors.Join(errs...)
}

// errLineTooLong is returned by readBoundedLine when a single newline-delimited
// frame exceeds maxWorkerLineBytes. The overrun is drained up to (and including)
// the terminating newline so the reader resynchronizes on the next frame instead
// of mis-parsing the tail as a fresh line.
var errLineTooLong = errors.New("worker output line exceeds limit")

// readBoundedLine reads a newline-terminated frame, capping buffered bytes at
// maxWorkerLineBytes. On a frame at or under the cap it behaves like
// ReadString('\n'). On an over-long frame it drains the remainder of the line
// and returns errLineTooLong with no usable data — callers recover for the
// affected context rather than OOMing or killing the worker.
func readBoundedLine(reader *bufio.Reader) (string, error) {
	var buf []byte
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return string(buf), err
		}
		if b == '\n' {
			return string(append(buf, b)), nil
		}
		if len(buf) >= maxWorkerLineBytes {
			// Drain to the next newline (or EOF) so the next read starts on a
			// clean frame boundary, then report the overrun.
			for b != '\n' {
				var derr error
				b, derr = reader.ReadByte()
				if derr != nil {
					return "", errLineTooLong
				}
			}
			return "", errLineTooLong
		}
		buf = append(buf, b)
	}
}

func (w *pythonSubprocessWorker) readLoop() {
	// Frames are newline-delimited JSON. A single legitimate frame (e.g. a large
	// base64 image payload from matplotlib) can exceed any fixed scanner cap, and
	// bufio.Scanner returns ErrTooLong and silently kills the loop in that case —
	// dropping the worker. Use a bufio.Reader with a large initial buffer and
	// readBoundedLine, which grows as needed up to maxWorkerLineBytes so a runaway
	// frame can't grow the buffer without bound ([B1] unbounded memory).
	reader := bufio.NewReaderSize(w.stdout, 64*1024)
	for {
		line, err := readBoundedLine(reader)
		if errors.Is(err, errLineTooLong) {
			// Over-long frame: the emit-side cap (per-field MAX_CHUNK_BYTES plus a
			// whole-line guard in repl_worker.py's _emit) makes this unreachable in
			// practice, but if it happens, surface an error for the current exec and
			// recover rather than OOM or silently kill the worker. The worker is
			// per-context so the blast radius is this one context.
			w.logger.Warn("dropping oversized worker chunk", slog.Int("limit", maxWorkerLineBytes))
			if w.onChunk != nil {
				w.onChunk(&WorkerChunk{
					SessionID: w.ctxID,
					Type:      ChunkTypeError,
					Name:      "WorkerProcessError",
					Value:     fmt.Sprintf("worker output frame exceeded %d bytes and was dropped", maxWorkerLineBytes),
				})
				w.onChunk(&WorkerChunk{
					SessionID: w.ctxID,
					Type:      ChunkTypeControl,
					Text:      ControlChunkTypeCompleted,
				})
			}
			// The aborted command's REAL terminal frame (the `completed`/
			// `interrupted` repl_worker.py always emits from its finally block) is
			// still queued behind the oversized line — plus possibly more of the
			// aborted command's trailing output. We have already synthesized a
			// terminal pair above, so if we returned to the normal loop those stale
			// frames would be mis-attributed to the NEXT queued command on this
			// context. Drain the subprocess stdout until that real terminal frame so
			// the next command starts on a clean boundary.
			if drainErr := w.drainToTerminal(reader); drainErr != nil {
				if !errors.Is(drainErr, io.EOF) {
					w.logger.Debug("worker readLoop ended draining after oversized chunk", slog.String("error", drainErr.Error()))
				}
				return
			}
			continue
		}
		if len(line) > 0 {
			var chunk WorkerChunk
			if jerr := json.Unmarshal([]byte(line), &chunk); jerr != nil {
				w.logger.Warn("ignoring malformed worker chunk", slog.String("error", jerr.Error()))
			} else if chunk.Type == "hostcall" {
				// bash() bridge request: route to the bash host and write the
				// result back to the worker's stdin. NOT forwarded as session
				// output. Handled off the read loop so a slow command can't delay
				// interrupt processing (see handleHostCall).
				var hc hostCall
				if uerr := json.Unmarshal([]byte(line), &hc); uerr == nil {
					go w.handleHostCall(&hc)
				}
			} else {
				chunk.SessionID = w.ctxID
				if w.onChunk != nil {
					w.onChunk(&chunk)
				}
			}
		}
		if err != nil {
			if !errors.Is(err, io.EOF) {
				w.logger.Debug("worker readLoop ended", slog.String("error", err.Error()))
			}
			return
		}
	}
}

// drainToTerminal discards frames from the subprocess stdout until (and
// including) the aborted command's real terminal control frame — the
// `completed`/`interrupted` chunk repl_worker.py always emits from its finally
// block. It is called only on the oversized-frame recovery path, after a
// synthetic terminal pair has already been delivered for the aborted command, to
// prevent the aborted command's trailing frames from leaking into the NEXT
// queued command on this context. Drained frames are NOT forwarded to onChunk
// (the exec is already terminated). Oversized frames encountered while draining
// are skipped too, so a runaway tail can't wedge the drain. Returns a read error
// (including io.EOF) if the stream ends before a terminal frame is seen.
func (w *pythonSubprocessWorker) drainToTerminal(reader *bufio.Reader) error {
	for {
		line, err := readBoundedLine(reader)
		if errors.Is(err, errLineTooLong) {
			// Another oversized frame in the aborted command's tail — keep draining.
			continue
		}
		if len(line) > 0 {
			var chunk WorkerChunk
			if jerr := json.Unmarshal([]byte(line), &chunk); jerr == nil &&
				chunk.Type == ChunkTypeControl &&
				(chunk.Text == ControlChunkTypeCompleted || chunk.Text == ControlChunkTypeInterrupted) {
				// Real terminal frame consumed — the stream is resynchronized.
				return nil
			}
		}
		if err != nil {
			return err
		}
	}
}

func (w *pythonSubprocessWorker) waitLoop() {
	err := w.cmd.Wait()
	w.active.Store(false)
	close(w.done)
	if err != nil {
		w.logger.Debug("python worker exited with error", slog.String("error", err.Error()))
		// Synthesize a worker-process error so the in-flight command (if any) terminates.
		if w.onChunk != nil {
			w.onChunk(&WorkerChunk{
				SessionID: w.ctxID,
				Type:      ChunkTypeError,
				Name:      "WorkerProcessError",
				Value:     err.Error(),
			})
			w.onChunk(&WorkerChunk{
				SessionID: w.ctxID,
				Type:      ChunkTypeControl,
				Text:      ControlChunkTypeCompleted,
			})
		}
	}
}
