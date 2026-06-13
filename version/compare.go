package version

// Compare compares two versions. Returns -1 if 'a' is less than 'b', 0 if 'a' is equal to 'b', and 1 if 'a' is greater than 'b'.
func Compare(a, b *Version) int {
	if a.Major > b.Major {
		return 1
	} else if a.Major < b.Major {
		return -1
	}

	if a.Minor > b.Minor {
		return 1
	} else if a.Minor < b.Minor {
		return -1
	}

	if a.Patch > b.Patch {
		return 1
	} else if a.Patch < b.Patch {
		return -1
	}

	return 0
}

// CompareStrings compares two version strings. Uses the same logic as the `Compare` function.
func CompareStrings(a, b string) (int, error) {
	av, err := FromString(a)
	if err != nil {
		return 0, err
	}
	bv, err := FromString(b)
	if err != nil {
		return 0, err
	}
	return Compare(av, bv), nil
}
