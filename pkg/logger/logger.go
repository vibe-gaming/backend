package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func Init(level string) {
	zLevel := getLevel(level)

	const (
		timestamp  = "@timestamp"
		severity   = "level"
		loggerName = "logger"
		caller     = "caller"
		message    = "message"
		stacktrace = "stacktrace"
	)

	logger = zap.New(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(zapcore.EncoderConfig{
				TimeKey:        timestamp,
				LevelKey:       severity,
				NameKey:        loggerName,
				CallerKey:      caller,
				MessageKey:     message,
				StacktraceKey:  stacktrace,
				LineEnding:     zapcore.DefaultLineEnding,
				EncodeLevel:    zapcore.LowercaseLevelEncoder,
				EncodeTime:     zapcore.ISO8601TimeEncoder,
				EncodeDuration: zapcore.SecondsDurationEncoder,
				EncodeCaller:   zapcore.ShortCallerEncoder,
			}),
			zapcore.AddSync(os.Stdout),
			zLevel,
		),
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	)
}

func getLevel(level string) zap.AtomicLevel {
	zLevel := zapcore.InfoLevel

	switch level {
	case "info":
		zLevel = zapcore.InfoLevel
	case "warn":
		zLevel = zapcore.WarnLevel
	case "error":
		zLevel = zapcore.ErrorLevel
	case "debug":
		zLevel = zapcore.DebugLevel
	}

	return zap.NewAtomicLevelAt(zLevel)
}

func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	logger.Fatal(msg, fields...)
}

func Sugar() *zap.SugaredLogger {
	return logger.Sugar()
}

func Logger() *zap.Logger {
	return logger
}
