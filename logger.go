package gin

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// LogFormatterParams is the structure any formatter will be handed when time to log comes
type LogFormatterParams struct {
	Request *http.Request
	// TimeStamp shows the time after the server returns a response.
	TimeStamp time.Time
	// StatusCode is HTTP Response Status code.
	StatusCode int
	// Latency is how much time the server spent to processing the request.
	Latency time.Duration
	// ClientIP equals ClientIP refers to the client IP address.
	ClientIP string
	// Method is the HTTP Request Method.
	Method string
	// Path is a path the client requested.
	Path string
	// ErrorMessage is the error message that was set in the context.
	ErrorMessage string
	// BodySize is the size of the Response Body.
	BodySize int
	// Keys are the keys set on the request's context.
	Keys map[string]interface{}
}

// LogFormatter gives the signature of the formatter function passed to LoggerWithFormatter
type LogFormatter func(params LogFormatterParams) string

// LoggerConfig defines the config for Logger middleware.
type LoggerConfig struct {
	// Formatter is the function that logs the request.
	// Optional. Default is defaultLogFormatter
	Formatter LogFormatter

	// Output is a writer where the logs will be written.
	// Optional. Default is gin.DefaultWriter
	Output io.Writer

	// SkipPaths is a url path array which logs are not written.
	// Optional.
	SkipPaths []string
}

// Logger instances a Logger middleware that will write the logs to gin.DefaultWriter.
// By default gin.DefaultWriter = os.Stdout.
func Logger() HandlerFunc {
	return LoggerWithConfig(LoggerConfig{})
}

// LoggerWithWriter instance a Logger middleware with the specified writer buffer.
// Example: os.Stdout, a file opened in write mode, a socket...
func LoggerWithWriter(out io.Writer, notlogged ...string) HandlerFunc {
	return LoggerWithConfig(LoggerConfig{
		Output:    out,
		SkipPaths: notlogged,
	})
}

// LoggerWithConfig instance a Logger middleware with config.
func LoggerWithConfig(conf LoggerConfig) HandlerFunc {
	formatter := conf.Formatter
	if formatter == nil {
		formatter = defaultLogFormatter
	}

	out := conf.Output
	if out == nil {
		out = DefaultWriter
	}

	var skip map[string]struct{}

	if length := len(conf.SkipPaths); length > 0 {
		skip = make(map[string]struct{}, length)

		for _, path := range conf.SkipPaths {
			skip[path] = struct{}{}
		}
	}

	return func(c *Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log only when path is not being skipped
		if _, ok := skip[path]; !ok {
			param := LogFormatterParams{
				Request: c.Request,
				Keys:    c.Keys,
			}

			// Stop timer
			param.TimeStamp = time.Now()
			param.Latency = param.TimeStamp.Sub(start)

			param.ClientIP = c.ClientIP()
			param.Method = c.Request.Method

			statusCode := c.Writer.Status()
			if statusCode == http.StatusOK && c.writermem.Status() != http.StatusOK {
				statusCode = c.writermem.Status()
			} else if statusCode == http.StatusOK && c.IsAborted() {
				// If the context was aborted but status is still 200, check if a status was set in writermem
				if c.writermem.Status() != http.StatusOK {
					statusCode = c.writermem.Status()
				}
			}
			param.StatusCode = statusCode

			param.ErrorMessage = c.Errors.ByType(ErrorTypePrivate).String()

			param.BodySize = c.Writer.Size()

			if raw != "" {
				path = path + "?" + raw
			}

			param.Path = path

			fmt.Fprint(out, formatter(param))
		}
	}
}

func defaultLogFormatter(param LogFormatterParams) string {
	var statusColor, methodColor, resetColor string
	if param.Keys != nil {
		// Support color logging if terminal supports it
	}
	return fmt.Sprintf("[GIN] %v |%s %3d %s| %13v | %15s |%s %-7s %s %#v\n%s",
		param.TimeStamp.Format("2006/01/02 - 15:04:05"),
		statusColor, param.StatusCode, resetColor,
		param.Latency,
		param.ClientIP,
		methodColor, param.Method, resetColor,
		param.Path,
		param.ErrorMessage,
	)
}