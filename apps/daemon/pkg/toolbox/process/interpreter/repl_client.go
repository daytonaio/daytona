// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package interpreter

import (
	"bufio"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

//go:embed repl_worker.py
var pythonWorkerScript string

// Info returns the current session information
func (s *InterpreterSession) Info() InterpreterSessionInfo {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.info
}

// enqueueAndExecute enqueues a job and processes jobs FIFO ensuring single execution at a time
func (s *InterpreterSession) enqueueAndExecute(code string, envs map[string]string, timeout time.Duration, ws *websocket.Conn) {
    s.mu.Lock()
    if s.queue == nil {
        s.queue = make(chan execJob, 128)
        go s.processQueue()
    }
    s.mu.Unlock()

    job := execJob{code: code, envs: envs, timeout: timeout, ws: ws}
    s.queue <- job
}

func (s *InterpreterSession) processQueue() {
    for job := range s.queue {
        // Attach the job's websocket as the sole client for the duration (non-blocking)
        if job.ws != nil {
            go s.attachWebSocket(job.ws)
        }
        _, _ = s.executeCodeWithEnvs(job.code, job.envs, job.timeout)
        // Close client connection at the end of this job to free the slot for next job
        s.closeAllClients()
    }
}

func (s *InterpreterSession) executeCodeWithEnvs(code string, envs map[string]string, timeout time.Duration) (*CommandExecution, error) {
    cmdID := uuid.NewString()
    execution := &CommandExecution{ID: cmdID, Code: code, Status: "running", StartedAt: time.Now()}
    s.commandMu.Lock()
    s.activeCommand = execution
    s.commandMu.Unlock()

    workerCmd := WorkerCommand{ID: cmdID, Cmd: "exec", Code: code, Envs: envs}
    if err := s.sendCommand(workerCmd); err != nil {
        execution.Status = "error"
        now := time.Now()
        execution.EndedAt = &now
        execution.Error = &Error{Name: "CommunicationError", Value: err.Error()}
        return execution, err
    }

    // Wait completion or timeout (0 = no timeout)
    resultChan := make(chan bool, 1)
    go func() {
        for {
            time.Sleep(50 * time.Millisecond)
            s.commandMu.Lock()
            status := s.activeCommand.Status
            s.commandMu.Unlock()
            if status != "running" {
                resultChan <- true
                return
            }
        }
    }()

    var timeoutC <-chan time.Time
    if timeout > 0 {
        timer := time.NewTimer(timeout)
        defer timer.Stop()
        timeoutC = timer.C
    }

    select {
    case <-resultChan:
        s.commandMu.Lock()
        result := s.activeCommand
        s.activeCommand = nil
        s.commandMu.Unlock()
        return result, nil
    case <-timeoutC:
        // SIGINT then grace
        if s.cmd != nil && s.cmd.Process != nil {
            _ = s.cmd.Process.Signal(syscall.SIGINT)
        }
        graceful := time.NewTimer(gracePeriod)
        defer graceful.Stop()
        select {
        case <-resultChan:
            s.commandMu.Lock()
            result := s.activeCommand
            s.activeCommand = nil
            s.commandMu.Unlock()
            return result, nil
        case <-graceful.C:
            if s.cmd != nil && s.cmd.Process != nil {
                _ = s.cmd.Process.Kill()
            }
            s.commandMu.Lock()
            if s.activeCommand != nil {
                s.activeCommand.Status = "interrupted"
                now := time.Now()
                s.activeCommand.EndedAt = &now
                result := s.activeCommand
                s.activeCommand = nil
                s.commandMu.Unlock()
                return result, errors.New("execution timeout")
            }
            s.commandMu.Unlock()
            return execution, errors.New("execution timeout")
        }
    }
}

// start initializes and starts the Python worker process
func (s *InterpreterSession) start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Already running?
    if s.info.Active && s.cmd != nil && s.stdin != nil {
		return nil
	}

    // Allow restart if previously used but not active

	ctx, cancel := context.WithCancel(context.Background())
	s.ctx = ctx
	s.cancel = cancel

    // Create (or reuse) a single worker script file per session
    tempDir := os.TempDir()
    workerPath := filepath.Join(tempDir, fmt.Sprintf("daytona_repl_worker_%s.py", s.info.ID))
    if st, statErr := os.Stat(workerPath); statErr != nil {
        if os.IsNotExist(statErr) {
            if err := os.WriteFile(workerPath, []byte(pythonWorkerScript), workerScriptPerms); err != nil {
                cancel()
                return fmt.Errorf("failed to create worker script: %w", err)
            }
        } else {
            cancel()
            return fmt.Errorf("failed to stat worker script: %w", statErr)
        }
    } else {
        // Ensure executable perms if file already exists
        _ = os.Chmod(workerPath, workerScriptPerms)
        _ = st // silence unused if not needed by linter
    }
	s.workerPath = workerPath

    // Start Python worker process
    pyCmd := detectPythonCommand()
    cmd := exec.CommandContext(ctx, pyCmd, workerPath)
	cmd.Dir = s.info.Cwd

    // Set only OS environment variables (no session-level envs)
    cmd.Env = os.Environ()

    // Get stdin/stdout pipes
    stdin, err := cmd.StdinPipe()
    if err != nil {
		cancel()
		os.Remove(workerPath)
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		stdin.Close()
		os.Remove(workerPath)
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

    // Capture worker process stderr into a buffer and mirror to stdout as error result
    // We still set cmd.Stderr = os.Stderr so daemon logs have diagnostics, but errors
    // from the worker are converted to a structured error when the process ends.
    cmd.Stderr = os.Stderr

	// Start the process
	if err := cmd.Start(); err != nil {
		cancel()
		stdin.Close()
		stdout.Close()
		os.Remove(workerPath)
		return fmt.Errorf("failed to start Python worker: %w", err)
	}

	s.cmd = cmd
	s.stdin = stdin
	s.stdout = stdout
	s.info.Active = true

    log.Debugf("Started interpreter session %s with PID %d", s.info.ID, s.cmd.Process.Pid)

    // Start reading worker output
    go s.workerReadLoop()

    // Monitor process exit
    go s.monitorProcess()

	return nil
}

// detectPythonCommand attempts to find a working python interpreter
func detectPythonCommand() string {
    candidates := []string{"python3", "python"}
    for _, c := range candidates {
        if _, err := exec.LookPath(c); err == nil {
            return c
        }
    }
    // Fallback to python3
    return "python3"
}

// sendCommand sends a command to the Python worker
func (s *InterpreterSession) sendCommand(cmd WorkerCommand) error {
	s.mu.Lock()
	stdin := s.stdin
	s.mu.Unlock()

	if stdin == nil {
		return errors.New("worker stdin not available")
	}

    data, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("failed to marshal command: %w", err)
	}

	data = append(data, '\n')
	_, err = stdin.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write command: %w", err)
	}

	return nil
}


// workerReadLoop reads messages from the Python worker
func (s *InterpreterSession) workerReadLoop() {
	scanner := bufio.NewScanner(s.stdout)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024) // Allow large lines

	for scanner.Scan() {
		line := scanner.Text()
        
        // New streaming protocol: line is a chunk with type
        var chunk map[string]any
        if err := json.Unmarshal([]byte(line), &chunk); err != nil {
            log.Errorf("Failed to parse worker chunk: %v", err)
            continue
        }
        s.handleChunk(chunk)
	}

	if err := scanner.Err(); err != nil {
		log.Errorf("Error reading from worker: %v", err)
	}
}


// handleChunk processes streaming chunks from the Python worker
func (s *InterpreterSession) handleChunk(chunk map[string]any) {
    chunkType, _ := chunk["type"].(string)
    
    // Update internal command state for certain chunk types
    if chunkType == "error" {
        s.commandMu.Lock()
        if s.activeCommand != nil {
            s.activeCommand.Status = "error"
            now := time.Now()
            s.activeCommand.EndedAt = &now
            s.activeCommand.Error = &Error{
                Name:      chunk["name"].(string),
                Value:     chunk["value"].(string),
                Traceback: chunk["traceback"].(string),
            }
        }
        s.commandMu.Unlock()
    } else if chunkType == "control" {
        // Handle completion signals
        controlText := getStringFromChunk(chunk, "text")
        s.commandMu.Lock()
        if s.activeCommand != nil {
            switch controlText {
            case "completed":
                s.activeCommand.Status = "ok"
                now := time.Now()
                s.activeCommand.EndedAt = &now
            case "interrupted":
                s.activeCommand.Status = "interrupted"
                now := time.Now()
                s.activeCommand.EndedAt = &now
            case "error_completed":
                // Error status already set by error chunk, just mark as ended
                if s.activeCommand.Status == "running" {
                    s.activeCommand.Status = "error"
                }
                now := time.Now()
                s.activeCommand.EndedAt = &now
            case "exit":
                s.activeCommand.Status = "exit"
                now := time.Now()
                s.activeCommand.EndedAt = &now
            }
        }
        s.commandMu.Unlock()
    }
    
    // Stream directly to WebSocket client
    msg := &OutputMessage{
        Type:      chunkType,
        Text:      getStringFromChunk(chunk, "text"),
        Name:      getStringFromChunk(chunk, "name"),
        Value:     getStringFromChunk(chunk, "value"),
        Traceback: getStringFromChunk(chunk, "traceback"),
    }
    if artifact, ok := chunk["artifact"].(map[string]any); ok {
        msg.Artifact = artifact
    }
    s.broadcastOutput(msg)
}

// Helper function to safely extract strings from chunk map
func getStringFromChunk(chunk map[string]any, key string) string {
    if val, ok := chunk[key].(string); ok {
        return val
    }
    return ""
}

// monitorProcess monitors the worker process and cleans up on exit
func (s *InterpreterSession) monitorProcess() {
    err := s.cmd.Wait()

	s.mu.Lock()
	s.info.Active = false
    sessionID := s.info.ID
	s.mu.Unlock()

    if err != nil {
        // Surface worker process error to the current active command if any
        s.commandMu.Lock()
        if s.activeCommand != nil && s.activeCommand.Status == "running" {
            s.activeCommand.Status = "error"
            now := time.Now()
            s.activeCommand.EndedAt = &now
            s.activeCommand.Error = &Error{
                Name:  "WorkerProcessError",
                Value: err.Error(),
            }
        }
        s.commandMu.Unlock()
        log.Debugf("Interpreter session %s process exited with error: %v", sessionID, err)
    } else {
        log.Debugf("Interpreter session %s process exited normally", sessionID)
    }

    // Keep worker script for reuse; do not remove

	// Close all WebSocket clients
	s.closeAllClients()

    // Single-session manager; no-op
}

// shutdown gracefully shuts down the worker
func (s *InterpreterSession) shutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.info.Active {
		return
	}

	// Send shutdown command
	shutdownCmd := WorkerCommand{
		ID:  uuid.NewString(),
		Cmd: "shutdown",
	}
	_ = s.sendCommand(shutdownCmd)

	// Give it a moment to exit gracefully
	time.AfterFunc(1*time.Second, func() {
		if s.cancel != nil {
			s.cancel()
		}
	})
}

// kill forcefully terminates the worker
func (s *InterpreterSession) kill() {
	s.mu.Lock()
	
	if !s.info.Active {
		s.mu.Unlock()
		return
	}

    sessionID := s.info.ID

	if s.cancel != nil {
		s.cancel()
	}
	if s.stdin != nil {
		_ = s.stdin.Close()
		s.stdin = nil
	}
	if s.stdout != nil {
		_ = s.stdout.Close()
		s.stdout = nil
	}
	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
	}
	s.info.Active = false
	s.mu.Unlock()

    // Keep worker script for reuse; do not remove

	// Close all WebSocket clients
	s.closeAllClients()

    // Single-session manager; no-op

	log.Debugf("Interpreter session %s killed", sessionID)
}


