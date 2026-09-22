// pkg/logger/logger.go
package logging

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger(release bool, logLevel string) *zap.Logger {
	level := zap.NewAtomicLevelAt(parseLogLevel(logLevel))

	file, err := os.OpenFile(getLogFilePath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Sprintf("failed to open log file: %v", err))
	}

	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(file),
		level,
	)

	var stdoutEncoder zapcore.Encoder
	if release {
		stdoutEncoder = zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	} else {
		stdoutEncoder = zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	}
	stdoutCore := zapcore.NewCore(stdoutEncoder, zapcore.AddSync(os.Stdout), level)

	core := zapcore.NewTee(stdoutCore, fileCore)
	return zap.New(core, zap.AddCaller())
}

func parseLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel // default to info if invalid
	}
}

func getLogFilePath() string {
	const logDir = "logs"

	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic(fmt.Sprintf("failed to create log directory: %v", err))
	}

	return filepath.Join(logDir, "app.log")
}
