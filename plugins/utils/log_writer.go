package utils

import (
	log "github.com/sirupsen/logrus"
)

type DebugLogWriter struct{}

func (w *DebugLogWriter) Write(p []byte) (n int, err error) {
	log.Debug(string(p))
	return len(p), nil
}
