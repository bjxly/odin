package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// L 全局 SugaredLogger，Init 之前使用 no-op 兜底避免空指针。
var L *zap.SugaredLogger

func init() {
	L = zap.NewNop().Sugar()
}

// Init 初始化全局日志器。
// level: debug / info / warn / error（大小写不敏感）
// format: json 或 console（大小写不敏感）
func Init(level, format string) {
	var lvl zapcore.Level
	if err := lvl.UnmarshalText([]byte(strings.ToLower(strings.TrimSpace(level)))); err != nil {
		lvl = zapcore.InfoLevel
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "ts"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderCfg.EncodeDuration = zapcore.StringDurationEncoder

	var encoder zapcore.Encoder
	if strings.EqualFold(strings.TrimSpace(format), "json") {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	} else {
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	}

	core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), lvl)
	core = zapcore.NewTee(core, zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		zapcore.AddSync(os.Stderr),
		zapcore.ErrorLevel,
	))

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(0))
	L = logger.Sugar()
}

// Sync 刷新缓冲区，程序退出前应调用。
func Sync() {
	if L != nil {
		_ = L.Sync()
	}
}

// Debug 输出 debug 级别日志。
func Debug(args ...interface{}) { L.Debug(args...) }

// Info 输出 info 级别日志。
func Info(args ...interface{}) { L.Info(args...) }

// Warn 输出 warn 级别日志。
func Warn(args ...interface{}) { L.Warn(args...) }

// Error 输出 error 级别日志。
func Error(args ...interface{}) { L.Error(args...) }

// Debugf 输出格式化 debug 日志。
func Debugf(tpl string, args ...interface{}) { L.Debugf(tpl, args...) }

// Infof 输出格式化 info 日志。
func Infof(tpl string, args ...interface{}) { L.Infof(tpl, args...) }

// Warnf 输出格式化 warn 日志。
func Warnf(tpl string, args ...interface{}) { L.Warnf(tpl, args...) }

// Errorf 输出格式化 error 日志。
func Errorf(tpl string, args ...interface{}) { L.Errorf(tpl, args...) }
