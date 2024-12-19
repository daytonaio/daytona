package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/api/types"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type LogReader struct {
	sync.Mutex
	position      *types.LogPosition
	retryCount    int
	maxRetries    int
	retryDelay    time.Duration
	activeProfile *config.Profile
	workspaceId   string
}

func NewLogReader(activeProfile *config.Profile, workspaceId string) *LogReader {
	return &LogReader{
		position:      &types.LogPosition{},
		maxRetries:    types.MaxReconnectAttempts,
		retryDelay:    types.InitialRetryDelay,
		activeProfile: activeProfile,
		workspaceId:   workspaceId,
	}
}

func (r *LogReader) SetMaxRetries(max int) {
	r.maxRetries = max
}

func (r *LogReader) ReadWorkspaceLogs(ctx context.Context, projectNames []string, follow bool, showWorkspaceLogs bool, from *time.Time) error {
	for {
		posJSON, _ := json.Marshal(r.position)
		query := fmt.Sprintf("retry=true&position=%s", url.QueryEscape(string(posJSON)))

		ws, _, err := GetWebsocketConn(ctx,
			fmt.Sprintf("/log/workspace/%s", r.workspaceId),
			r.activeProfile, &query)

		if err != nil {
			if r.retryCount >= r.maxRetries {
				return fmt.Errorf("max retries reached: %w", err)
			}
			log.Debug("Connection failed, retrying: ", err)
			time.Sleep(r.retryDelay)
			r.retryDelay = min(r.retryDelay*2, types.MaxRetryDelay)
			r.retryCount++
			continue
		}

		// Reset retry counters on successful connection
		r.retryCount = 0
		r.retryDelay = types.InitialRetryDelay

		if err := r.handleLogStream(ctx, ws, from, projectNames, follow, showWorkspaceLogs); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
				log.Debug("Connection closed unexpectedly, retrying")
				continue
			}
			return err
		}

		if !follow {
			break
		}
	}

	return nil
}

func (r *LogReader) handleLogStream(ctx context.Context, ws *websocket.Conn, from *time.Time, projectNames []string, follow bool, showWorkspaceLogs bool) error {
	defer ws.Close()

	ws.SetPingHandler(func(string) error {
		return ws.WriteControl(websocket.PongMessage, []byte{}, time.Now().Add(time.Second))
	})

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			_, message, err := ws.ReadMessage()
			if err != nil {
				return err
			}

			// Try to parse as position update
			var pos types.LogPosition
			if err := json.Unmarshal(message, &pos); err == nil {
				r.Lock()
				r.position = &pos
				r.Unlock()
				continue
			}

			// Process log message
			if err := r.processLogMessage(message, from, projectNames, showWorkspaceLogs); err != nil {
				log.Debug("Error processing message: ", err)
			}
		}
	}
}

func (r *LogReader) processLogMessage(message []byte, from *time.Time, projectNames []string, showWorkspaceLogs bool) error {
	var logEntry struct {
		Time      string `json:"time"`
		ProjectID string `json:"project_id,omitempty"`
	}

	if err := json.Unmarshal(message, &logEntry); err != nil {
		return err
	}

	if from != nil {
		parsedTime, err := time.Parse(time.RFC3339Nano, logEntry.Time)
		if err != nil {
			return err
		}

		if parsedTime.Before(*from) {
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
