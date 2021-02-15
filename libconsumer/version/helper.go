package version

import "time"

// GetDefaultVersion returns the curent libconsumer version
func GetDefaultVersion() string {
	if qualifier == "" {
		return defaultConsumerVersion
	}

	return defaultConsumerVersion + "-" + qualifier
}

var (
	buildTime = "unknown"
	commit    = "unknown"
	qualifier = ""
)

// BuildTime exposes the compile-time build time information.
// It will represent the zero time instant if parsing fails.
func BuildTime() time.Time {
	t, err := time.Parse(time.RFC3339, buildTime)
	if err != nil {
		return time.Time{}
	}
	return t
}

// Commit exposes the compile-time hash.
func Commit() string {
	return commit
}
