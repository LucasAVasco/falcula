// Package version contains functions for working with this application's version.
package version

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// Version represents a version.
type Version struct {
	Major  uint
	Minor  uint
	Patch  uint
	Suffix string
}

// New returns a new version with the given version parts.
func New(major, minor, patch uint, suffix string) *Version {
	return &Version{
		Major:  major,
		Minor:  minor,
		Patch:  patch,
		Suffix: suffix,
	}
}

// FromString returns a new version from the given version string. It must be in the format "vMajor.Minor.Patch-Suffix". The first 'v' and
// the suffix are optional.
func FromString(version string) (*Version, error) {
	if version == "" {
		return nil, fmt.Errorf("version string must not be empty")
	}

	if version[0] == 'v' {
		version = version[1:]
	}

	split := strings.Split(version, ".")
	if len(split) < 3 {
		return nil, fmt.Errorf("invalid version string (must be in the format 'major.minor.patch'): %s", version)
	}

	major, err := strconv.Atoi(split[0])
	if err != nil {
		return nil, fmt.Errorf("error parsing major version: %w", err)
	}

	minor, err := strconv.Atoi(split[1])
	if err != nil {
		return nil, fmt.Errorf("error parsing minor version: %w", err)
	}

	// Parse the patch and suffix
	splitIndex := strings.IndexFunc(split[2], func(r rune) bool {
		return !unicode.IsDigit(r)
	})
	var patchString string
	var suffix string
	if splitIndex > 0 {
		patchString = split[2][:splitIndex]
		suffix = split[2][splitIndex:]
	} else {
		patchString = split[2]
		suffix = ""
	}

	patch, err := strconv.Atoi(patchString)
	if err != nil {
		return nil, fmt.Errorf("error parsing patch version: %w", err)
	}

	return New(uint(major), uint(minor), uint(patch), suffix), nil
}

// String returns the version string.
func (v *Version) String() string {
	return fmt.Sprintf("v%d.%d.%d-%s", v.Major, v.Minor, v.Patch, v.Suffix)
}

// Compare returns -1 if 'v' is less than 'other', 0 if 'v' is equal to 'other', and 1 if 'v' is greater than 'other'.
func (v *Version) Compare(other *Version) int {
	return Compare(v, other)
}

// CompareStrings compares the version string of 'v' with 'other'. Uses the same logic as the `Compare` function.
func (v *Version) CompareStrings(other string) (int, error) {
	return CompareStrings(v.String(), other)
}
