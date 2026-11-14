//go:build darwin || freebsd || netbsd || openbsd

package rawinput

import "golang.org/x/sys/unix"

const (
	ioctlGetTermios = unix.TIOCGETA
	ioctlSetTermios = unix.TIOCSETA
)
