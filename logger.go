package main

import (
	"context"

	sentryzap "github.com/getsentry/sentry-go/zap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(cfg *config) *zap.Logger {
	var baseLogger *zap.Logger
	if cfg.IsProd() {
		baseLogger = zap.Must(zap.NewProduction())
	} else {
		baseLogger = zap.Must(zap.NewDevelopment())
	}

	sentryCore := sentryzap.NewSentryCore(context.Background(), sentryzap.Option{
		Level: []zapcore.Level{
			zapcore.ErrorLevel,
		},
		AddCaller: true,
	})

	return zap.New(zapcore.NewTee(baseLogger.Core(), sentryCore))
}
