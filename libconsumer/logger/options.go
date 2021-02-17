package logger

// Option configures the logger package behavior.
type Option func(cfg *Config)

// WithLevel specifies the logger level.
func WithLevel(level Level) Option {
	return func(cfg *Config) {
		cfg.Level = level
	}
}
