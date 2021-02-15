package instance

import (
	"errors"
	"syscall"
)

var errNotImplemented = errors.New("not implemented on platform")

func setUmask(newmask int) error {
	syscall.Umask(newmask)
	return nil // the umask syscall always succeeds: http://man7.org/linux/man-pages/man2/umask.2.html#RETURN_VALUE
}
