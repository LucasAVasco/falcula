//go:build linux || darwin || freebsd || netbsd || openbsd

package rawinput

import (
	"errors"

	"golang.org/x/sys/unix"
)

// pollTimeoutMs is how often the reader checks if it must stop when there is no input
const pollTimeoutMs = 100

// makeCbreak disables the line buffering and the echo of the terminal. The output processing and the signals keep working. Returns a
// function that restores the previous configuration
func makeCbreak(fd int) (restore func(), err error) {
	old, err := unix.IoctlGetTermios(fd, ioctlGetTermios)
	if err != nil {
		return nil, err
	}

	cbreak := *old
	cbreak.Lflag &^= unix.ICANON | unix.ECHO
	cbreak.Cc[unix.VMIN] = 1
	cbreak.Cc[unix.VTIME] = 0
	if err := unix.IoctlSetTermios(fd, ioctlSetTermios, &cbreak); err != nil {
		return nil, err
	}

	return func() { _ = unix.IoctlSetTermios(fd, ioctlSetTermios, old) }, nil
}

// waitForAltT reads the file descriptor until Alt+T is pressed (returns true), the `stop` channel is closed or the input ends (returns
// false)
func waitForAltT(fd int, stop <-chan struct{}) bool {
	var detector altTDetector
	buffer := make([]byte, 64)
	fds := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}

	for {
		select {
		case <-stop:
			return false
		default:
		}

		ready, err := unix.Poll(fds, pollTimeoutMs)
		if errors.Is(err, unix.EINTR) {
			continue
		}
		if err != nil {
			return false
		}
		if ready == 0 {
			continue // Timeout: checks if it must stop
		}

		n, err := unix.Read(fd, buffer)
		if errors.Is(err, unix.EINTR) || errors.Is(err, unix.EAGAIN) {
			continue
		}
		if err != nil || n == 0 {
			return false // The input was closed
		}

		for _, b := range buffer[:n] {
			if detector.Feed(b) {
				return true
			}
		}
	}
}
