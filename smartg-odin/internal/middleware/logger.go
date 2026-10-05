package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"smartg-odin/internal/common/logger"
)

// TraceID 链路追踪中间件：优先复用请求头中的 X-Trace-Id，
// 否则生成新的 UUID，并写入 gin.Context 与响应头。
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		tid := c.GetHeader("X-Trace-Id")
		if tid == "" {
			tid = uuid.NewString()
		}
		c.Set(TraceIDKey, tid)
		c.Header("X-Trace-Id", tid)
		c.Next()
	}
}

// Logger 请求日志中间件：记录 method、path、status、latency、traceId 等信息。
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		tid := c.GetString(TraceIDKey)

		fields := []interface{}{
			"trace_id", tid,
			"method", c.Request.Method,
			"path", path,
			"query", query,
			"status", status,
			"latency", latency.String(),
			"latency_ms", latency.Milliseconds(),
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
			"errors", c.Errors.ByType(gin.ErrorTypePrivate).String(),
		}

		switch {
		case status >= 500:
			logger.L.Errorw("request", fields...)
		case status >= 400:
			logger.L.Warnw("request", fields...)
		default:
			logger.L.Infow("request", fields...)
		}
	}
}
