package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"smartg-odin/internal/common/errs"
	"smartg-odin/internal/common/logger"
)

// Recovery panic 恢复中间件：捕获处理过程中的 panic，
// 记录堆栈日志并返回统一错误码 50000。
// 若响应已写入（如客户端断开连接），则不再重复写入。
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				stack := string(debug.Stack())
				logger.L.Errorw("panic recovered",
					"trace_id", c.GetString(TraceIDKey),
					"error", r,
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"stack", stack,
				)

				// 已经写过响应体则不再写入，避免 superfluous WriteHeader 警告
				if !c.Writer.Written() {
					ErrorWithHTTP(c, http.StatusInternalServerError,
						errs.New(errs.CodeInternal, "internal server error"))
				}
				c.Abort()
			}
		}()
		c.Next()
	}
}
