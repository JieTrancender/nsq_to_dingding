package logger

import "testing"

func TestLogger(t *testing.T) {
	exerciseLogger := func() {
		log := NewLogger("exercise")
		log.Info("some message")
		log.Infof("some message with parameter x = %v, y = %v", 1, 2)
		log.Infow("some message", "x", 1, "y", 2)
		log.Infow("some message", Int("x", 1), Int("y", 2))
		log.Infow("some message with namespaced args", Namespace("metrics"), "x", 1, "y", 2)
		log.Infow("", "empty_message", true)

		// Add context.
		log.With("x", 1, "y", 2).Warn("logger with context")

		someStruct := struct {
			X int `json:"x"`
			Y int `json:"y"`
		}{1, 2}
		log.Infow("some message with struct value", "metrics", someStruct)
	}

	// TestingSetup()
	exerciseLogger()
}
