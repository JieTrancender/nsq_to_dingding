package consumer

import "errors"

// ErrGracefulExit is an error that signals to exit with a code of 0
var ErrGracefulExit = errors.New("graceful exit")
