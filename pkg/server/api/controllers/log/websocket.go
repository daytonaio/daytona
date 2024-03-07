package log

import (
	"context"
	"net/http"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/server/config"
	"github.com/daytonaio/daytona/pkg/server/db"
	"github.com/daytonaio/daytona/pkg/server/logs"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func writeToWs(ws *websocket.Conn, c chan []byte, errChan chan error) {
	for {
		err := ws.WriteMessage(websocket.TextMessage, <-c)
		if err != nil {
			errChan <- err
			break
		}
	}
}

func readLog(ginCtx *gin.Context, logFilePath *string) {
	followQuery := ginCtx.Query("follow")
	follow := followQuery == "true"

	ws, err := upgrader.Upgrade(ginCtx.Writer, ginCtx.Request, nil)
	if err != nil {
		log.Error(err)
		return
	}

	msgChannel := make(chan []byte)
	errChannel := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()
	go util.ReadLog(ctx, logFilePath, follow, msgChannel, errChannel)
	go writeToWs(ws, msgChannel, errChannel)

	go func() {
		err := <-errChannel
		if err != nil {
			log.Error(err)
		}
		ws.Close()
		cancel()
	}()

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			log.Error(err)
			cancel()
			break
		}
	}
}

func ReadServerLog(ginCtx *gin.Context) {
	readLog(ginCtx, logs.LogFilePath)
}

func ReadWorkspaceLog(ginCtx *gin.Context) {
	workspaceId := ginCtx.Param("workspaceId")

	workspace, err := db.FindWorkspaceByIdOrName(workspaceId)
	if err != nil {
		log.Error(err)
		return
	}

	workspaceLogFilePath, err := config.GetWorkspaceLogFilePath(workspace.Id)
	if err != nil {
		log.Error(err)
		return
	}

	readLog(ginCtx, &workspaceLogFilePath)
}
