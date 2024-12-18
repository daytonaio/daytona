package apiclient

import (
    "context"
    "encoding/json"
    "fmt"
    "net/url"
    "sync"
    "time"

    "github.com/gorilla/websocket"
    log "github.com/sirupsen/logrus"
)

type LogReader struct {
    sync.Mutex
    position    *LogPosition
    retryCount  int
    maxRetries  int
    retryDelay  time.Duration
    activeProfile *config.Profile
    workspaceId  string
}

func NewLogReader(activeProfile *config.Profile, workspaceId string) *LogReader {
    return &LogReader{
        position:      &LogPosition{},
        maxRetries:    maxReconnectAttempts,
        retryDelay:    initialRetryDelay,
        activeProfile: activeProfile,
        workspaceId:   workspaceId,
    }
}

func (r *LogReader) ReadWorkspaceLogs(ctx context.Context, showWorkspaceLogs bool, from *time.Time) error {
    for {
        posJSON, _ := json.Marshal(r.position)
        query := fmt.Sprintf("retry=true&position=%s", url.QueryEscape(string(posJSON)))
        
        ws, res, err := GetWebsocketConn(ctx, 
            fmt.Sprintf("/log/workspace/%s", r.workspaceId), 
            r.activeProfile, &query)
        
        if err != nil {
            if r.retryCount >= r.maxRetries {
                return fmt.Errorf("max retries reached: %w", err)
            }
            log.Debug("Connection failed, retrying: ", err)
            time.Sleep(r.retryDelay)
            r.retryDelay = min(r.retryDelay*2, maxRetryDelay)
            r.retryCount++
            continue
        }
        
        // Reset retry counters on successful connection
        r.retryCount = 0
        r.retryDelay = initialRetryDelay

        if err := r.handleLogStream(ctx, ws, from); err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
                log.Debug("Connection closed unexpectedly, retrying")
                continue
            }
            return err
        }
    }
}

func (r *LogReader) handleLogStream(ctx context.Context, ws *websocket.Conn, from *time.Time) error {
    defer ws.Close()

    // Setup ping handler
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
            var pos LogPosition
            if err := json.Unmarshal(message, &pos); err == nil {
                r.Lock()
                r.position = &pos
                r.Unlock()
                continue
            }

            // Process log message
            if from != nil {
                var logEntry logs.LogEntry
                if err := json.Unmarshal(message, &logEntry); err == nil {
                    parsedTime, err := time.Parse(time.RFC3339Nano, logEntry.Time)
                    if err == nil && (parsedTime.After(*from) || parsedTime.Equal(*from)) {
                        fmt.Println(string(message))
                    }
                }
            } else {
                fmt.Println(string(message))
            }
        }
    }
}