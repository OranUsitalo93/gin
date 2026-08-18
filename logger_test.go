package gin

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/assert"
)

func performRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestLoggerWithAbort(t *testing.T) {
	signature := ""
	buf := new(bytes.Buffer)
	
	router := New()
	// Attach logger with custom writer to capture output
	router.Use(LoggerWithWriter(buf))
	
	// Middleware that aborts the request
	router.Use(func(c *Context) {
		c.AbortWithStatus(http.StatusUnauthorized)
	})
	
	router.GET("/test", func(c *Context) {
		signature = "should not be executed"
	})

	w := performRequest(router, "GET", "/test")

	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Empty(t, signature)
	assert.Contains(t, buf.String(), "401")
	assert.NotContains(t, buf.String(), "200")
}