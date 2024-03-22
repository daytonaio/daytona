package logger

import (
	"io"
	"os"
	"path/filepath"
)

type projectLogger struct {
	logsDir     string
	workspaceId string
	projectName string
	logFile     *os.File
}

func (pl *projectLogger) Write(p []byte) (n int, err error) {
	if pl.logFile == nil {
		filePath := filepath.Join(pl.logsDir, pl.workspaceId, pl.projectName, "log")
		err = os.MkdirAll(filepath.Dir(filePath), 0755)
		if err != nil {
			return 0, err
		}

		logFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return 0, err
		}
		pl.logFile = logFile
	}

	return pl.logFile.Write(p)
}

func (pl *projectLogger) Close() error {
	if pl.logFile != nil {
		err := pl.logFile.Close()
		pl.logFile = nil
		return err
	}
	return nil
}

func GetProjectLogger(logsDir, workspaceId, projectName string) io.WriteCloser {
	return &projectLogger{workspaceId: workspaceId, logsDir: logsDir, projectName: projectName}
}
