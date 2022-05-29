package authentication

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gin-gonic/gin"
)

func TestHeaderAuthMiddleware(t *testing.T) {
	handler := HeaderAuthMiddleware()

	tests := map[string]struct {
		ctx       *gin.Context
		isAborted bool
	}{
		"empty context": {
			ctx:       newContext(map[string]string{}),
			isAborted: true,
		},
		"X-Token=test in context": {
			ctx: newContext(map[string]string{AuthTokenHeader: AuthTokenValue}),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			handler(test.ctx)
			assert.Equal(t, test.isAborted, test.ctx.IsAborted())
		})
	}
}

func newContext(headers map[string]string) *gin.Context {
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request, _ = http.NewRequest("GET", "/ping", nil)
	for key, value := range headers {
		ctx.Request.Header.Add(key, value)
	}
	return ctx
}
