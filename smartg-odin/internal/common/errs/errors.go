package errs

import "fmt"

// 统一业务错误码
const (
	CodeSuccess      = 0
	CodeParamError   = 40000
	CodeUnauthorized = 40100
	CodeForbidden    = 40300
	CodeNotFound     = 40400
	CodeTooManyReqs  = 42900
	CodeInternal     = 50000
	CodeDBError      = 50001
	CodeConnError    = 50002
	CodeQueryTimeout = 50003
	CodeConnRefused  = 50300
	CodeServiceDown  = 50301
)

// AppError 应用级错误，携带业务错误码与提示信息。
type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

// Error 实现 error 接口。
func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err != nil {
		return fmt.Sprintf("code=%d message=%s err=%v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("code=%d message=%s", e.Code, e.Message)
}

// Unwrap 暴露底层错误，支持 errors.Is / errors.As。
func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// New 创建一个新的 AppError。
func New(code int, msg string) *AppError {
	return &AppError{Code: code, Message: msg}
}

// Wrap 在底层错误之上包装为 AppError。
func Wrap(code int, msg string, err error) *AppError {
	return &AppError{Code: code, Message: msg, Err: err}
}

// From 将任意 error 转换为 AppError；若已经是 AppError 则原样返回。
func From(err error) *AppError {
	if err == nil {
		return nil
	}
	if ae, ok := err.(*AppError); ok {
		return ae
	}
	return Wrap(CodeInternal, err.Error(), err)
}

// ---- 便捷构造函数 ----

func ParamError(msg string) *AppError {
	if msg == "" {
		msg = "invalid parameter"
	}
	return New(CodeParamError, msg)
}

func Unauthorized(msg string) *AppError {
	if msg == "" {
		msg = "unauthorized"
	}
	return New(CodeUnauthorized, msg)
}

func Forbidden(msg string) *AppError {
	if msg == "" {
		msg = "forbidden"
	}
	return New(CodeForbidden, msg)
}

func NotFound(msg string) *AppError {
	if msg == "" {
		msg = "resource not found"
	}
	return New(CodeNotFound, msg)
}

func TooManyRequests(msg string) *AppError {
	if msg == "" {
		msg = "too many requests"
	}
	return New(CodeTooManyReqs, msg)
}

func Internal(msg string, err error) *AppError {
	if msg == "" {
		msg = "internal server error"
	}
	return Wrap(CodeInternal, msg, err)
}

func DBError(err error) *AppError {
	return Wrap(CodeDBError, "database error", err)
}

func ConnError(err error) *AppError {
	return Wrap(CodeConnError, "datasource connection error", err)
}

func QueryTimeout(err error) *AppError {
	return Wrap(CodeQueryTimeout, "query timeout", err)
}

func ConnRefused(err error) *AppError {
	return Wrap(CodeConnRefused, "connection refused", err)
}

func ServiceDown(err error) *AppError {
	return Wrap(CodeServiceDown, "service unavailable", err)
}
