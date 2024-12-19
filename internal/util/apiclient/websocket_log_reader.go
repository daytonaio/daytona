package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type LogPosition struct {
	Offset    int64     `json:"offset"`
	Timestamp time.Time `json:"timestamp"`
}

const (
	defaultMaxRetries = 5
	defaultRetryDelay = time.Second
	defaultMaxDelay   = 30 * time.Second
)

// LogReader handles reading logs with reconnection support
type LogReader struct {
	mu            sync.Mutex
	position      *LogPosition
	maxRetries    int
	retryDelay    time.Duration
	activeProfile *config.Profile
	workspaceId   string
}

// NewLogReader creates a new LogReader instance
func NewLogReader(profile *config.Profile, workspaceId string) *LogReader {
	return &LogReader{
		position:      &LogPosition{},
		maxRetries:    defaultMaxRetries,
		retryDelay:    defaultRetryDelay,
		activeProfile: profile,
		workspaceId:   workspaceId,
	}
}

func (r *LogReader) SetMaxRetries(max int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.maxRetries = max
}

func ReadWorkspaceLogs(ctx context.Context, profile *config.Profile, workspaceId string, projectNames []string, follow bool, showWorkspaceLogs bool, from *time.Time) error {
	r := NewLogReader(profile, workspaceId)
	retryCount := 0

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Create query parameters
			query := url.Values{}
			query.Set("retry", "true")
			if r.position != nil {
				posBytes, err := json.Marshal(r.position)
				if err != nil {
					return fmt.Errorf("failed to marshal position: %w", err)
				}
				query.Set("position", string(posBytes))
			}

			// Connect to websocket
			wsURL := fmt.Sprintf("/log/workspace/%s?%s", r.workspaceId, query.Encode())
			ws, _, err := GetWebsocketConn(ctx, wsURL, r.activeProfile, nil)
			if err != nil {
				if retryCount >= r.maxRetries {
					return fmt.Errorf("max retries reached: %w", err)
				}
				retryCount++
				time.Sleep(r.retryDelay)
				r.retryDelay = time.Duration(min(r.retryDelay.Nanoseconds()*2, defaultMaxDelay.Nanoseconds())) * time.Nanosecond
				continue
			}

			// Reset retry counters on successful connection
			retryCount = 0
			r.retryDelay = defaultRetryDelay

			if err := r.handleLogStream(ctx, ws, from, projectNames, follow, showWorkspaceLogs); err != nil {
				ws.Close()
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
					continue
				}
				return err
			}

			if !follow {
				return nil
			}
		}
	}
}

func (r *LogReader) handleLogStream(ctx context.Context, ws *websocket.Conn, from *time.Time, projectNames []string, _ bool, showWorkspaceLogs bool) error {
	defer ws.Close()

	// Set up ping handler
	ws.SetPingHandler(func(message string) error {
		deadline := time.Now().Add(time.Second)
		err := ws.WriteControl(websocket.PongMessage, []byte{}, deadline)
		if err != nil {
			log.Errorf("Failed to write pong message: %v", err)
			return err
		}
		return nil
	})

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			_, message, err := ws.ReadMessage()
			if err != nil {
				return err
			}

			// Try to parse as position update
			var pos LogPosition
			if err := json.Unmarshal(message, &pos); err == nil {
				r.mu.Lock()
				r.position = &pos
				r.mu.Unlock()
				continue
			}

			if err := r.processLogMessage(message, from, projectNames, showWorkspaceLogs); err != nil {
				log.Debugf("Error processing message: %v", err)
			}
		}
	}
}

// processLogMessage handles individual log messages
func (r *LogReader) processLogMessage(message []byte, from *time.Time, projectNames []string, showWorkspaceLogs bool) error {
	var logEntry struct {
		Time      string `json:"time"`
		ProjectID string `json:"project_id,omitempty"`
		Message   string `json:"message"`
	}

	if err := json.Unmarshal(message, &logEntry); err != nil {
		return fmt.Errorf("failed to unmarshal log entry: %w", err)
	}

	// Filter by time if specified
	if from != nil {
		timestamp, err := time.Parse(time.RFC3339, logEntry.Time)
		if err != nil {
			return fmt.Errorf("failed to parse timestamp: %w", err)
		}
		if timestamp.Before(*from) {
			return nil
		}
	}

	// Filter by project if needed
	if len(projectNames) > 0 && logEntry.ProjectID != "" {
		found := false
		for _, name := range projectNames {
			if name == logEntry.ProjectID {
				found = true
				break
			}
		}
		if !found && !showWorkspaceLogs {
			return nil
		}
	}

	fmt.Println(string(message))
	return nil
}

// min returns the minimum of two int64 values
func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
