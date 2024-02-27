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

		if len(ctx.Errors) > 0 {
			log.WithFields(log.Fields{
				"method":  reqMethod,
				"URI":     reqUri,
				"status":  statusCode,
				"latency": latencyTime,
				"error":   ctx.Errors.String(),
			}).Error("API ERROR")
			ctx.JSON(statusCode, gin.H{"error": ctx.Errors[0].Err.Error()})
		} else {
			log.WithFields(log.Fields{
				"method":  reqMethod,
				"URI":     reqUri,
				"status":  statusCode,
				"latency": latencyTime,
			}).Info("API REQUEST")
		}

		ctx.Next()
	}
}
