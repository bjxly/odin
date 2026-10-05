package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"smartg-odin/internal/common/errs"
)

// TraceIDKey 存储在 gin.Context 中的 traceId 键名。
const TraceIDKey = "trace_id"

// Response 统一响应结构。
type Response struct {
	Code      int         `json:"code"`
	Data      interface{} `json:"data,omitempty"`
	Message   string      `json:"message"`
	Success   bool        `json:"success"`
	Timestamp int64       `json:"timestamp"`
	TraceID   string      `json:"trace_id"`
}

// PageResult 分页结果。
type PageResult struct {
	Records interface{} `json:"records"`
	Total   int64       `json:"total"`
	Current int         `json:"current"`
	Size    int         `json:"size"`
}

// traceID 从上下文读取 traceId。
func traceID(c *gin.Context) string {
	if c == nil {
		return ""
	}
	return c.GetString(TraceIDKey)
}

// Success 返回成功响应（HTTP 200）。
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:      errs.CodeSuccess,
		Data:      data,
		Message:   "ok",
		Success:   true,
		Timestamp: time.Now().UnixMilli(),
		TraceID:   traceID(c),
	})
}

// SuccessWithMessage 返回带自定义提示的成功响应。
func SuccessWithMessage(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusOK, Response{
		Code:      errs.CodeSuccess,
		Data:      data,
		Message:   message,
		Success:   true,
		Timestamp: time.Now().UnixMilli(),
		TraceID:   traceID(c),
	})
}

// SuccessPage 返回分页成功响应。
func SuccessPage(c *gin.Context, records interface{}, total int64, current, size int) {
	Success(c, PageResult{
		Records: records,
		Total:   total,
		Current: current,
		Size:    size,
	})
}

// Error 返回业务错误响应，HTTP 状态码统一 200，错误通过 code 区分。
func Error(c *gin.Context, appErr *errs.AppError) {
	ErrorWithHTTP(c, http.StatusOK, appErr)
}

// ErrorWithHTTP 返回业务错误响应，并指定 HTTP 状态码。
func ErrorWithHTTP(c *gin.Context, httpStatus int, appErr *errs.AppError) {
	if appErr == nil {
		appErr = errs.New(errs.CodeInternal, "internal server error")
	}
	c.JSON(httpStatus, Response{
		Code:      appErr.Code,
		Data:      nil,
		Message:   appErr.Message,
		Success:   false,
		Timestamp: time.Now().UnixMilli(),
		TraceID:   traceID(c),
	})
}
