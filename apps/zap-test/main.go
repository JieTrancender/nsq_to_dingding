package main

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func f1(url string) {
	testFunc(fmt.Sprintf("Failed to fetch URL: %s", url))
}

func testFunc(msg string, fields ...zap.Field) {
	zap.L().Warn(msg, fields...)
}

func main() {
	// logger := zap.NewExample()
	// logger, _ := zap.NewDevelopment()
	// logger, _ := zap.NewProduction(zap.AddCaller(), zap.AddCallerSkip(1))
	logger, _ := zap.NewProduction(zap.AddCaller(), zap.AddStacktrace(zapcore.WarnLevel))
	zap.ReplaceGlobals(logger)

	logger2 := logger.With(zap.String("services", "manager-api"))

	url := "http://example.org/api"
	logger.Info("failed to fetch URL",
		zap.String("url", url),
		zap.Int("attempt", 3),
		zap.Duration("backoff", time.Second),
	)

	sugar := logger.Sugar()
	sugar.Infow("failed to fetch URL",
		"url", url,
		"attempt", 3,
		"backoff", time.Second,
	)
	sugar.Infof("Failed to fetch URL: %s", url)

	logger2.Info("failed to fetch URL")

	f1(url)

	// _ = logger.Sync()
}
