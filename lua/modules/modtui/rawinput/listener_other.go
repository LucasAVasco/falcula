//go:build !(linux || darwin || freebsd || netbsd || openbsd)

package rawinput

func makeCbreak(int) (func(), error) {
	return nil, ErrUnsupported
}

func waitForAltT(int, <-chan struct{}) bool {
	return false
}
