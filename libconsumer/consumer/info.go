package consumer

// Info stores a consumer instance meta data.
type Info struct {
	Consumer string // The actual beat's name
	Version  string // The beat version. Defaults to the libconsumer version when an implementation does not set a version.
}
