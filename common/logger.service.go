package common

import (
	"context"

	"go.uber.org/zap"
)

var logger *zap.Logger

type loggerContextKey struct{}

func InitializeLogger() {
	if IsDev() {
		initializeDevLogger()
	} else {
		initializeProdLogger()
	}
	if logger != nil {
		logger.Info("logger initialized")
	}
}

func initializeProdLogger() {
	// TODO: implement
	// this needs to out put to file and stdout
	// and also needs to be rotated
	// and also needs to be sent to a log aggregator (maybe)
}

func initializeDevLogger() {
	logger = zap.Must(zap.NewDevelopment())
}

func GetLogger() *zap.Logger {
	if logger == nil {
		InitializeLogger()
	}
	return logger
}

func LoggerFromCtx(ctx context.Context) *zap.Logger {
	ctxLogger, ok := ctx.Value(loggerContextKey{}).(*zap.Logger)
	if !ok || ctxLogger == nil {
		return GetLogger()
	}
	return ctxLogger
}

func SetLoggerToCtx(ctx context.Context, logger *zap.Logger) context.Context {
	ctxLogger := LoggerFromCtx(ctx)
	if ctxLogger == logger {
		return ctx
	}
	return context.WithValue(ctx, loggerContextKey{}, logger)
}
