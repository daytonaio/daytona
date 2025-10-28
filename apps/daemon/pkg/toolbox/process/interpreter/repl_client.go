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
		// Attach the websocket client
		if job.ws != nil {
			go s.attachWebSocket(job.ws)
		}
		
		// Execute the code
		_, err := s.executeCodeWithEnvs(job.code, job.envs, job.timeout)
		
		// Close the websocket with appropriate code
		if err != nil && err.Error() == "execution timeout" {
			s.closeClient(WebSocketCloseTimeout, "Execution timeout")
		} else {
			s.closeClient(websocket.CloseNormalClosure, "execution completed")
		}
	}
}

// closeClient closes the WebSocket client with specified close code
func (s *InterpreterSession) closeClient(code int, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.client == nil {
		return
	}
	
	// Send close frame with appropriate code
	closeMsg := websocket.FormatCloseMessage(code, message)
	_ = s.client.conn.SetWriteDeadline(time.Now().Add(writeWait))
	_ = s.client.conn.WriteMessage(websocket.CloseMessage, closeMsg)
	
	// Close the client
	s.client.close()
	s.client = nil
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

	// Wait for completion or timeout (0 = no timeout)
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
		// Send SIGINT for graceful interruption
		if s.cmd != nil && s.cmd.Process != nil {
			_ = s.cmd.Process.Signal(syscall.SIGINT)
		}
		
		// Wait for grace period
		graceful := time.NewTimer(gracePeriod)
		defer graceful.Stop()
		
		select {
		case <-resultChan:
			// Completed during grace period
			s.commandMu.Lock()
			result := s.activeCommand
			s.activeCommand = nil
			s.commandMu.Unlock()
			return result, nil
			
		case <-graceful.C:
			// Grace period expired - force kill
			if s.cmd != nil && s.cmd.Process != nil {
				_ = s.cmd.Process.Kill()
			}
			
			s.commandMu.Lock()
			if s.activeCommand != nil {
				s.activeCommand.Status = "timeout"
				now := time.Now()
				s.activeCommand.EndedAt = &now
				s.activeCommand.Error = &Error{
					Name:  "TimeoutError",
					Value: "Execution timeout - code took too long to execute",
				}
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
	s.done = make(chan struct{})

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
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()

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
				Name:      getStringFromChunk(chunk, "name"),
				Value:     getStringFromChunk(chunk, "value"),
				Traceback: getStringFromChunk(chunk, "traceback"),
			}
		}
		s.commandMu.Unlock()
	} else if chunkType == "control" {
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

	// Stream to WebSocket client
	s.broadcastOutput(&OutputMessage{
		Type:      chunkType,
		Text:      getStringFromChunk(chunk, "text"),
		Name:      getStringFromChunk(chunk, "name"),
		Value:     getStringFromChunk(chunk, "value"),
		Traceback: getStringFromChunk(chunk, "traceback"),
		Artifact:  getArtifactFromChunk(chunk),
	})
}

// Helper functions
func getStringFromChunk(chunk map[string]any, key string) string {
	if val, ok := chunk[key].(string); ok {
		return val
	}
	return ""
}

func getArtifactFromChunk(chunk map[string]any) map[string]any {
	if artifact, ok := chunk["artifact"].(map[string]any); ok {
		return artifact
	}
	return nil
}

// monitorProcess monitors the worker process and cleans up on exit
func (s *InterpreterSession) monitorProcess() {
	err := s.cmd.Wait()

	s.mu.Lock()
	s.info.Active = false
	sessionID := s.info.ID
	done := s.done
	s.mu.Unlock()

	// Notify waiters that the process has exited
	if done != nil {
		close(done)
	}

	if err != nil {
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

	// Close WebSocket client if any
	s.closeClient(websocket.CloseGoingAway, "worker process ended")
}

// shutdown gracefully shuts down the worker
func (s *InterpreterSession) shutdown() {
	s.mu.Lock()
	
	if !s.info.Active {
		s.mu.Unlock()
		return
	}

	// Get references while we have the lock
	sessionID := s.info.ID
	cancel := s.cancel
	cmd := s.cmd
	done := s.done
	s.mu.Unlock()
	
	// Send shutdown command
	shutdownCmd := WorkerCommand{
		ID:  uuid.NewString(),
		Cmd: "shutdown",
	}
	_ = s.sendCommand(shutdownCmd)

	// Wait for process to exit (monitorProcess will close the done channel)
	if done != nil {
		select {
		case <-done:
			// Process exited gracefully
			log.Debugf("Interpreter session %s shut down gracefully", sessionID)
		case <-time.After(2 * time.Second):
			// Timeout - force kill
			log.Debugf("Interpreter session %s shutdown timeout, force killing", sessionID)
			if cancel != nil {
				cancel()
			}
			if cmd != nil && cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
			// Wait a bit more for kill to take effect
			time.Sleep(100 * time.Millisecond)
		}
	}
	
	// Close WebSocket client
	s.closeClient(websocket.CloseGoingAway, "session shutdown")
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

	// Close WebSocket client
	s.closeClient(websocket.CloseGoingAway, "session killed")

	log.Debugf("Interpreter session %s killed", sessionID)
}
