package logger

// Environment indicates the enviroment the logger is supped to be run in.
// The default logger configuration may be different for different environments.
type Environment int

const (
	// DefaultEnvironment is used if the environment the process runs in is not known.
	DefaultEnvironment Environment = iota

	// SystemdEnvironment indicates that the process is started and managed by systemd.
	SystemdEnvironment

	// InvalidEnviroment indicates that the enviroment name given is unknown or invalid.
	InvalidEnviroment
)

// String returns the string representation the configured environment
func (v Environment) String() string {
	switch v {
	case DefaultEnvironment:
		return "default"
	case SystemdEnvironment:
		return "systemd"
	default:
		return "<invalid>"
	}
}
