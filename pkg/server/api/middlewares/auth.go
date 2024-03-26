package middlewares

import (
	"errors"
	"strings"

	"github.com/daytonaio/daytona/pkg/server/auth"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		bearerToken := ctx.GetHeader("Authorization")
		if bearerToken == "" {
			ctx.AbortWithError(401, errors.New("unauthorized"))
			return
		}

		token := extractToken(bearerToken)
		if token == "" {
			ctx.AbortWithError(401, errors.New("unauthorized"))
			return
		}

		if !auth.IsValidApiKey(token) {
			ctx.AbortWithError(401, errors.New("unauthorized"))
		}

		ctx.Next()
	}
}

func extractToken(bearerToken string) string {
	if !strings.HasPrefix(bearerToken, "Bearer ") {
		return ""
	}

	return strings.TrimPrefix(bearerToken, "Bearer ")
}
