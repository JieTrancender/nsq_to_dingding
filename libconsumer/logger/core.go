package logger

import (
	"flag"
	golog "log"
	"sync/atomic"
	"unsafe"

	"go.uber.org/zap"
)

var (
	_log          unsafe.Pointer // Pointer to a coreLogger. Access via atomic.LoadPointer.
	_defaultGoLog = golog.Writer()
)

func init() {
	storeLogger(&coreLogger{
		selectors:  map[string]struct{}{},
		rootLogger: zap.NewNop(),
		logger:     newLogger(zap.NewNop(), ""),
	})
}

type coreLogger struct {
	selectors  map[string]struct{}
	rootLogger *zap.Logger // Root logger without any options configured.
	logger     *Logger     // Logger that is the basis for all logp.Loggers.
}

// Configure configures the logger package.
func Configure(cfg Config) error {
	// return ConfigureWithOutputs(cfg)
	return nil
}

func loadLogger() *coreLogger {
	p := atomic.LoadPointer(&_log)
	return (*coreLogger)(p)
}

func storeLogger(l *coreLogger) {
	if old := loadLogger(); old != nil {
		old.rootLogger.Sync()
	}
	atomic.StorePointer(&_log, unsafe.Pointer(l))
}

// DevelopmentSetup configures the logger in development mode at debug level.
// By default the output goes to stderr.
func DevelopmentSetup(options ...Option) error {
	cfg := Config{
		Level: DebugLevel,
		// isStderr:    true,
		IsStderr:    true,
		development: true,
		addCaller:   true,
	}
	for _, apply := range options {
		apply(&cfg)
	}
	return Configure(cfg)
}

// TestingSetup configures logger by calling DevelopmentSetup if and only if verbose testing is enabled (as in 'go test -v').
func TestingSetup(options ...Option) error {
	// Use the flag to avoid a dependency on the testing package.
	f := flag.Lookup("test.v")
	if f != nil && f.Value.String() == "true" {
		return DevelopmentSetup(options...)
	}

	return nil
}
