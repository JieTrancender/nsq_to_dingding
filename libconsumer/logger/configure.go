package logger

import "time"

// Config contains the configuration options for the logger.
// To create a config from a common.Config use logger/config.Build.
type Config struct {
	Consumer  string   `config:",ignore"`   // Name of the consumer.
	JSON      bool     `config:"json"`      // Write logs as JSON.
	Level     Level    `config:"level"`     // Logging level (error, warning, info, debug)
	Selectors []string `config:"selectors"` // Selectors for debug level logger.

	toIODiscard bool
	// isStderr    bool `config:"isstderr" yaml: "isstderr"`
	IsStderr bool `config:"stderr"`
	ToFiles  bool `config:"to_files" yaml:"to_files"`

	Files FileConfig `config:"files"`

	enviroment  Environment
	addCaller   bool // Adds package and line number info to messages.
	development bool // Controls how DPanic behaves.
}

// FileConfig contains the configuration options for the file output.
type FileConfig struct {
	Path            string        `config:"path" yaml:"path"`
	Name            string        `config:"name" yaml:"name"`
	MaxSize         uint          `config:"rotateeverybytes" yaml:"rotateeverybytes" validate:"min=1"`
	Permissions     uint32        `config:"permissions"`
	Interval        time.Duration `config:"interval"`
	RotateOnStartup bool          `config:"rotateonstartup"`
	RedirectStderr  bool          `config:"redirect_stderr" yaml:"redirect_stderr"`
}

const defaultLevel = InfoLevel

// DefaultConfig returns the default config options for a given enviroment witch the Consumer is supposed to be run within.
func DefaultConfig(enviroment Environment) Config {
	return Config{
		Level: defaultLevel,
		Files: FileConfig{
			MaxSize:         10 * 1024 * 1024,
			Permissions:     0600,
			Interval:        0,
			RotateOnStartup: true,
		},
		enviroment: enviroment,
		addCaller:  true,
	}
}
