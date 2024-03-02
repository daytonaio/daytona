package log

import (
	"bufio"
	"context"
	"net/http"
	"os/exec"

	"github.com/daytonaio/daytona/pkg/server/config"
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

func readLog(ctx context.Context, filePath *string, follow bool, c chan []byte, errChan chan error) {
	if filePath == nil {
		return
	}

	ctxCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	tailCmd := exec.CommandContext(ctxCancel, "tail", "-n", "+1")
	if follow {
		tailCmd.Args = append(tailCmd.Args, "-f")
	}
	tailCmd.Args = append(tailCmd.Args, *filePath)

	reader, err := tailCmd.StdoutPipe()
	if err != nil {
		errChan <- err
		return
	}
	scanner := bufio.NewScanner(reader)
	go func() {
		for scanner.Scan() {
			c <- scanner.Bytes()
		}
	}()

	err = tailCmd.Start()
	if err != nil {
		errChan <- err
		return
	}

	errChan <- tailCmd.Wait()
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

func ReadServerLog(ginCtx *gin.Context) {
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
	go readLog(ctx, logs.LogFilePath, follow, msgChannel, errChannel)
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

func ReadWorkspaceLog(ginCtx *gin.Context) {
	workspaceId := ginCtx.Param("workspaceId")

	projectLogFilePath, err := config.GetWorkspaceLogFilePath(workspaceId)
	if err != nil {
		log.Error(err)
		return
	}

	ws, err := upgrader.Upgrade(ginCtx.Writer, ginCtx.Request, nil)
	if err != nil {
		log.Error(err)
		return
	}

	msgChannel := make(chan []byte)
	errChannel := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()
	go readLog(ctx, &projectLogFilePath, true, msgChannel, errChannel)
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
