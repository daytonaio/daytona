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

	common_errors "github.com/daytonaio/common-go/pkg/errors"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

//go:embed repl_worker.py
var pythonWorkerScript string

// Info returns the current context information
func (c *Context) Info() ContextInfo {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.info
}

// enqueueAndExecute enqueues a job and processes jobs FIFO ensuring single execution at a time
func (c *Context) enqueueAndExecute(code string, envs map[string]string, timeout time.Duration, ws *websocket.Conn) {
	c.mu.Lock()
	if c.queue == nil {
		c.queue = make(chan execJob, 128)
		go c.processQueue()
	}
	c.mu.Unlock()

	job := execJob{code: code, envs: envs, timeout: timeout, ws: ws}
	c.queue <- job
}

func (c *Context) processQueue() {
	for job := range c.queue {
		if job.ws != nil {
			go c.attachWebSocket(job.ws)
		}

		result, err := c.executeCode(job.code, job.envs, job.timeout)

		if err != nil && common_errors.IsRequestTimeoutError(err) || result.Status == CommandStatusTimeout {
			c.closeClient(WebSocketCloseTimeout, "")
		} else {
			c.closeClient(websocket.CloseNormalClosure, "")
		}
	}
}

// closeClient closes the WebSocket client with specified close code
func (c *Context) closeClient(code int, message string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client == nil {
		return
	}

	c.client.requestClose(code, message)
	c.client = nil
}

// executeCode executes code in the interpreter context
func (c *Context) executeCode(code string, envs map[string]string, timeout time.Duration) (*CommandExecution, error) {
	cmdID := uuid.NewString()
	execution := &CommandExecution{
		ID:        cmdID,
		Code:      code,
		Status:    CommandStatusRunning,
		StartedAt: time.Now(),
	}

	c.commandMu.Lock()
	c.activeCommand = execution
	c.commandMu.Unlock()

	workerCmd := WorkerCommand{ID: cmdID, Code: code, Envs: envs}
	err := c.sendCommand(workerCmd)
	if err != nil {
		execution.Status = CommandStatusError
		now := time.Now()
		execution.EndedAt = &now
		execution.Error = &Error{Name: "CommunicationError", Value: err.Error()}
		return execution, err
	}

	resultChan := make(chan bool, 1)
	go func() {
		for {
			time.Sleep(50 * time.Millisecond)
			c.commandMu.Lock()
			if c.activeCommand == nil || c.activeCommand.Status != CommandStatusRunning {
				c.commandMu.Unlock()
				resultChan <- true
				return
			}
			c.commandMu.Unlock()
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
		c.commandMu.Lock()
		result := c.activeCommand
		c.activeCommand = nil
		c.commandMu.Unlock()
		return result, nil

	case <-timeoutC:
		if c.cmd != nil && c.cmd.Process != nil {
			_ = c.cmd.Process.Signal(syscall.SIGINT)
		}

		graceful := time.NewTimer(gracePeriod)
		defer graceful.Stop()

		select {
		case <-resultChan:
			c.commandMu.Lock()
			result := c.activeCommand
			c.activeCommand = nil
			c.commandMu.Unlock()
			return result, nil

		case <-graceful.C:
			if c.cmd != nil && c.cmd.Process != nil {
				_ = c.cmd.Process.Kill()
			}

			c.commandMu.Lock()
			if c.activeCommand != nil {
				c.activeCommand.Status = CommandStatusTimeout
				now := time.Now()
				c.activeCommand.EndedAt = &now
				c.activeCommand.Error = &Error{
					Name:  "TimeoutError",
					Value: "Execution timeout - code took too long to execute",
				}
				result := c.activeCommand
				c.activeCommand = nil
				c.commandMu.Unlock()
				return result, common_errors.NewRequestTimeoutError(fmt.Errorf("execution timeout"))
			}
			c.commandMu.Unlock()
			return execution, common_errors.NewRequestTimeoutError(fmt.Errorf("execution timeout"))
		}
	}
}

// start initializes and starts the Python worker process
func (c *Context) start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Already running?
	if c.info.Active && c.cmd != nil && c.stdin != nil {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	c.ctx = ctx
	c.cancel = cancel

	// Create (or reuse) a single shared worker script file
	tempDir := os.TempDir()
	workerPath := filepath.Join(tempDir, "daytona_repl_worker.py")

	// Check if worker file exists, if not create it
	if _, err := os.Stat(workerPath); os.IsNotExist(err) {
		err := os.WriteFile(workerPath, []byte(pythonWorkerScript), workerScriptPerms)
		if err != nil {
			cancel()
			return fmt.Errorf("failed to create worker script: %w", err)
		}
	}

	c.workerPath = workerPath

	// Start Python worker process
	pyCmd := detectPythonCommand()
	cmd := exec.CommandContext(ctx, pyCmd, workerPath)
	cmd.Dir = c.info.Cwd
	cmd.Env = os.Environ()

	// Get stdin/stdout pipes
	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		stdin.Close()
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	cmd.Stderr = os.Stderr

	// Start the process
	err = cmd.Start()
	if err != nil {
		cancel()
		stdin.Close()
		stdout.Close()
		return fmt.Errorf("failed to start Python worker: %w", err)
	}

	c.cmd = cmd
	c.stdin = stdin
	c.stdout = stdout
	c.info.Active = true
	c.done = make(chan struct{})

	c.logger.Debug("Started interpreter context", "contextId", c.info.ID, "pid", c.cmd.Process.Pid)

	// Start reading worker output
	go c.workerReadLoop()

	// Monitor process exit
	go c.monitorProcess()

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
func (c *Context) sendCommand(cmd WorkerCommand) error {
	c.mu.Lock()
	stdin := c.stdin
	c.mu.Unlock()

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
func (c *Context) workerReadLoop() {
	scanner := bufio.NewScanner(c.stdout)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()

		var chunk map[string]any
		err := json.Unmarshal([]byte(line), &chunk)
		if err != nil {
			c.logger.Error("Failed to parse worker chunk", "error", err)
			continue
		}
		c.handleChunk(chunk)
	}

	err := scanner.Err()
	if err != nil {
		c.logger.Error("Error reading from worker", "error", err)
	}
}

// handleChunk processes streaming chunks from the Python worker
func (c *Context) handleChunk(chunk map[string]any) {
	// Extract all fields at the beginning
	chunkType := getStringFromChunk(chunk, "type")
	text := getStringFromChunk(chunk, "text")
	name := getStringFromChunk(chunk, "name")
	value := getStringFromChunk(chunk, "value")
	traceback := getStringFromChunk(chunk, "traceback")

	// Update internal command state for certain chunk types
	switch chunkType {
	case ChunkTypeError:
		c.commandMu.Lock()
		if c.activeCommand != nil {
			c.activeCommand.Status = CommandStatusError
			now := time.Now()
			c.activeCommand.EndedAt = &now
			c.activeCommand.Error = &Error{
				Name:      name,
				Value:     value,
				Traceback: traceback,
			}
		}
		c.commandMu.Unlock()
	case ChunkTypeControl:
		c.commandMu.Lock()
		if c.activeCommand != nil {
			switch text {
			case ControlChunkTypeCompleted:
				// Only set to OK if no error occurred (status would be Error already)
				if c.activeCommand.Status == CommandStatusRunning {
					c.activeCommand.Status = CommandStatusOK
				}
				now := time.Now()
				c.activeCommand.EndedAt = &now
			case ControlChunkTypeInterrupted:
				c.activeCommand.Status = CommandStatusTimeout
				now := time.Now()
				c.activeCommand.EndedAt = &now
			}
		}
		c.commandMu.Unlock()
		return
	}

	// Stream to WebSocket client
	c.emitOutput(&OutputMessage{
		Type:      chunkType,
		Text:      text,
		Name:      name,
		Value:     value,
		Traceback: traceback,
	})
}

// Helper functions
func getStringFromChunk(chunk map[string]any, key string) string {
	if val, ok := chunk[key].(string); ok {
		return val
	}
	return ""
}

// monitorProcess monitors the worker process and cleans up on exit
func (c *Context) monitorProcess() {
	err := c.cmd.Wait()

	c.mu.Lock()
	c.info.Active = false
	contextID := c.info.ID
	done := c.done
	c.mu.Unlock()

	// Notify waiters that the process has exited
	if done != nil {
		close(done)
	}

	if err != nil {
		c.commandMu.Lock()
		if c.activeCommand != nil && c.activeCommand.Status == CommandStatusRunning {
			c.activeCommand.Status = CommandStatusError
			now := time.Now()
			c.activeCommand.EndedAt = &now
			c.activeCommand.Error = &Error{
				Name:  "WorkerProcessError",
				Value: err.Error(),
			}
		}
		c.commandMu.Unlock()
		c.logger.Error("Interpreter context process exited with error", "contextId", contextID, "error", err)
	} else {
		c.logger.Debug("Interpreter context process exited normally", "contextId", contextID)
	}

	// Close WebSocket client if any
	c.closeClient(websocket.CloseGoingAway, "worker process ended")
}

// shutdown gracefully shuts down the worker
func (c *Context) shutdown() {
	c.mu.Lock()

	if !c.info.Active {
		c.mu.Unlock()
		return
	}

	// Get references while we have the lock
	contextID := c.info.ID
	cancel := c.cancel
	cmd := c.cmd
	done := c.done
	queue := c.queue
	c.mu.Unlock()

	// Close the queue to exit processQueue goroutine and prevent new jobs
	if queue != nil {
		close(queue)
		c.queue = nil
	}

	// Send SIGTERM to trigger immediate graceful shutdown (not queued)
	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Signal(syscall.SIGTERM)
	}

	// Wait for process to exit (monitorProcess will close the done channel)
	if done != nil {
		select {
		case <-done:
			// Process exited gracefully
			c.logger.Debug("Interpreter context shut down gracefully", "contextId", contextID)
		case <-time.After(2 * time.Second):
			// Timeout - force kill
			c.logger.Debug("Interpreter context shutdown timeout, force killing", "contextId", contextID)
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
	c.closeClient(websocket.CloseGoingAway, "context shutdown")
}
