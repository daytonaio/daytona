package middlewares

import (
	"time"

	"github.com/gin-gonic/gin"

	log "github.com/sirupsen/logrus"
)

func LoggingMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		startTime := time.Now()
		ctx.Next()
		endTime := time.Now()
		latencyTime := endTime.Sub(startTime)
		reqMethod := ctx.Request.Method
		reqUri := ctx.Request.RequestURI
		statusCode := ctx.Writer.Status()

		log.WithFields(log.Fields{
			"method":  reqMethod,
			"URI":     reqUri,
			"status":  statusCode,
			"latency": latencyTime,
		}).Info("API REQUEST")

		ctx.Next()
	}
}
