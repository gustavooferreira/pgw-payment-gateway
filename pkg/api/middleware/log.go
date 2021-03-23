package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gustavooferreira/pgw-payment-gateway-service/pkg/core/log"
)

// GinReqLogger returns a gin.HandlerFunc (middleware) that logs requests.
//
// Requests with errors are logged at the Error level
// Requests without errors are logged at the Info level
//
// It receives:
//   1. A logger
//   2. A time package format string (e.g. time.RFC3339).
//   3. A string specifying the message to print.
//   3. A string specifying the message type (if empty, don't create this field).
//
// Note: This code was copied from https://github.com/gin-contrib/zap with some modifications.
func GinReqLogger(logger log.Logger, timeFormat string, msg string, msgType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		// some evil middlewares modify these values
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		if len(c.Errors) > 0 {
			// Append error field if this is an erroneous request.
			for _, e := range c.Errors.Errors() {
				logger.Error(e, nil)
			}
		} else {
			fields := log.FieldsMap{
				"status":     c.Writer.Status(),
				"method":     c.Request.Method,
				"path":       path,
				"query":      query,
				"ip":         c.ClientIP(),
				"user-agent": c.Request.UserAgent(),
				"latency":    latency.Seconds(),
			}

			// Check if X-Request-ID header exists, if yes, log it too
			requestID := c.Request.Header.Get("X-Request-ID")

			if requestID != "" {
				fields["requestid"] = requestID
			}

			if msgType != "" {
				fields["type"] = msgType
			}

			logger.Info(msg, log.Fields(fields))
		}
	}
}
