package logger

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"gotest.tools/assert"
)

func TestLoggerWithOptions(t *testing.T) {
	core1, observed1 := observer.New(zapcore.DebugLevel)
	core2, observed2 := observer.New(zapcore.DebugLevel)

	logger1 := NewLogger("logger", zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewTee(core, core1)
	}))
	logger2 := logger1.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewTee(core, core2)
	}))

	logger1.Info("Hello logger1")
	logger2.Info("Hello logger1 and logger2")

	observedEntries1 := observed1.All()
	require.Len(t, observedEntries1, 2)
	assert.Equal(t, "Hello logger1", observedEntries1[0].Message)
	assert.Equal(t, "Hello logger1 and logger2", observedEntries1[1].Message)

	observedEntries2 := observed2.All()
	require.Len(t, observedEntries2, 1)
	assert.Equal(t, "Hello logger1 and logger2", observedEntries2[0].Message)
}
