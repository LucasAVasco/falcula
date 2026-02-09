// Package arch is a collection of functions related to the system platform and architecture
package arch

import "runtime"

// GetCurrentPlatform returns the current platform in the format used by Docker. Example: 'linux/amd64'
func GetCurrentPlatform() string {
	return runtime.GOOS + "/" + runtime.GOARCH
}
