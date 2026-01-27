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
	var cfg zap.Config

	if release {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
	}

	logFile := getLogFilePath()

	cfg.OutputPaths = []string{
		"stdout",
		logFile,
	}
	cfg.ErrorOutputPaths = []string{
		"stderr",
		logFile,
	}

	level := parseLogLevel(logLevel)
	cfg.Level = zap.NewAtomicLevelAt(level)

	logger, err := cfg.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build logger: %v", err))
	}

	return logger
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
