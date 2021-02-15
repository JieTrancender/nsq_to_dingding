package v2

// Plugin describes an input type.
type Plugin struct {
	Name string

	// Info contains a short description of the input type.
	Info string

	// Doc contains an optional longer description
	Doc string

	// Manager must be configured. The manager is used to create the inputs.
	// Manager InputManager
}
