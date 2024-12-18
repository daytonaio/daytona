package log

import (
    "context"
    "encoding/json"
    "errors"
    "io"
    "net/http"
    "sync"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    log "github.com/sirupsen/logrus"
)

const (
    maxReconnectAttempts = 5
    initialRetryDelay    = 1 * time.Second
    maxRetryDelay       = 30 * time.Second
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

type WebSocketReader struct {
    sync.Mutex
    position    *LogPosition
    ws          *websocket.Conn
    readFunc    func(context.Context, io.Reader, bool, chan<- []byte, chan<- error)
    writeFunc   func(*websocket.Conn, []byte) error
}

func NewWebSocketReader(ws *websocket.Conn, readFunc func(context.Context, io.Reader, bool, chan<- []byte, chan<- error), writeFunc func(*websocket.Conn, []byte) error) *WebSocketReader {
    return &WebSocketReader{
        ws:        ws,
        readFunc:  readFunc,
        writeFunc: writeFunc,
        position:  &LogPosition{},
    }
}

func (r *WebSocketReader) UpdatePosition(offset int64, line string) {
    r.Lock()
    defer r.Unlock()
    r.position.Offset = offset
    r.position.Timestamp = time.Now()
    r.position.LastLine = line
}

func readLog(ginCtx *gin.Context, logReader io.Reader, readFunc func(context.Context, io.Reader, bool, chan<- []byte, chan<- error), writeFunc func(*websocket.Conn, []byte) error) {
    positionQuery := ginCtx.Query("position")
    var lastPosition *LogPosition
    if positionQuery != "" {
        var err error
        lastPosition, err = UnmarshalPosition(positionQuery)
        if err != nil {
            log.Warn("Failed to parse position: ", err)
        }
    }

    ws, err := upgrader.Upgrade(ginCtx.Writer, ginCtx.Request, nil)
    if err != nil {
        log.Error(err)
        return
    }

    wsReader := NewWebSocketReader(ws, readFunc, writeFunc)
    if lastPosition != nil {
        wsReader.position = lastPosition
    }

    defer func() {
        closeErr := websocket.CloseNormalClosure
        if err != nil && !errors.Is(err, io.EOF) {
            closeErr = websocket.CloseInternalServerErr
        }
        ws.WriteControl(websocket.CloseMessage, 
            websocket.FormatCloseMessage(closeErr, ""), 
            time.Now().Add(time.Second))
        ws.Close()
    }()

    ctx, cancel := context.WithCancel(ginCtx.Request.Context())
    defer cancel()

    msgChan := make(chan []byte)
    errChan := make(chan error)
    heartbeatTicker := time.NewTicker(15 * time.Second)
    defer heartbeatTicker.Stop()

    // Start reading from log
    go readFunc(ctx, logReader, true, msgChan, errChan)

    // Handle websocket read messages (client messages)
    go func() {
        for {
            _, _, err := ws.ReadMessage()
            if err != nil {
                if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
                    log.Debug("Websocket read error: ", err)
                }
                errChan <- err
                return
            }
        }
    }()

    for {
        select {
        case <-ctx.Done():
            return
        case msg := <-msgChan:
            wsReader.UpdatePosition(int64(len(msg)), string(msg))
            if err := writeFunc(ws, msg); err != nil {
                log.Debug("Write error: ", err)
                return
            }
        case err := <-errChan:
            if err != nil {
                if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
                    // Send current position before closing
                    posJSON, _ := json.Marshal(wsReader.position)
                    writeFunc(ws, posJSON)
                }
                return
            }
        case <-heartbeatTicker.C:
            // Send heartbeat and position
            if err := ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second)); err != nil {
                log.Debug("Heartbeat failed: ", err)
                return
            }
            posJSON, _ := json.Marshal(wsReader.position)
            if err := writeFunc(ws, posJSON); err != nil {
                log.Debug("Position update failed: ", err)
                return
            }
        }
    }
}