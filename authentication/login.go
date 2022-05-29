package authentication

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HeaderAuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		header := HeaderAuth{
			Name:  AuthTokenHeader,
			Value: ctx.Request.Header.Get(AuthTokenHeader),
		}
		if header.Login() {
			log.Printf("Auth success: %s:%s\n", header.Name, header.Value)
			ctx.Next()
		} else {
			log.Printf("Auth failure: %s:%s\n", header.Name, header.Value)
			ctx.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}
