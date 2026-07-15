package logging

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger = *zap.SugaredLogger

func NewLogger() Logger {
	config := zap.NewProductionConfig()
	config.Encoding = "console"
	config.DisableStacktrace = true
	config.DisableCaller = true
	config.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.Level.SetLevel(getLogLevel())
	logger, _ := config.Build()

	defer logger.Sync() // flushes buffer, if any
	return logger.Sugar()
}

func getLogLevel() zapcore.Level {
	level := os.Getenv("LOG_LEVEL")

	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	}

	return zapcore.InfoLevel
}
