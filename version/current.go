package version

import (
	"fmt"
	"runtime/debug"
)

// CurrentVersionString is the current version of the application as a string. If it can not be determined, it will be empty.
var CurrentVersionString string

// CurrentVersion is the current version of the application. If it can not be determined, it will be nil.
var CurrentVersion *Version

// GetCurrentVersionString returns the current version of the application as a string.
func GetCurrentVersionString() (string, error) {
	info, ok := debug.ReadBuildInfo()
	if ok {
		return info.Main.Version, nil
	}

	return "", fmt.Errorf("error getting current version string")
}

func init() {
	var err error
	CurrentVersionString, err = GetCurrentVersionString()
	if err != nil {
		CurrentVersionString = ""
		return
	}

	CurrentVersion, err = FromString(CurrentVersionString)
	if err != nil {
		CurrentVersion = nil
		return
	}
}
