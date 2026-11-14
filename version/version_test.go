package version

import "testing"

func TestFromString_ValidVersions_AreParsed(t *testing.T) {
	tests := []struct {
		input               string
		major, minor, patch uint
		suffix              string
	}{
		{"1.2.3", 1, 2, 3, ""},
		{"v1.2.3", 1, 2, 3, ""},
		{"v10.20.30-rc1", 10, 20, 30, "-rc1"},
		{"v0.0.10-0.20250101120000-abcdef123456", 0, 0, 10, "-0.20250101120000-abcdef123456"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			v, err := FromString(tt.input)
			if err != nil {
				t.Fatalf("FromString(%q) returned error: %v", tt.input, err)
			}
			if v.Major != tt.major || v.Minor != tt.minor || v.Patch != tt.patch || v.Suffix != tt.suffix {
				t.Fatalf("FromString(%q) = %+v, want %d.%d.%d%s", tt.input, v, tt.major, tt.minor, tt.patch, tt.suffix)
			}
		})
	}
}

func TestFromString_InvalidVersions_ReturnError(t *testing.T) {
	for _, input := range []string{"", "v", "1.2", "a.2.3", "1.b.3", "1.2.c", "1.2.-rc1", "(devel)"} {
		t.Run(input, func(t *testing.T) {
			if v, err := FromString(input); err == nil {
				t.Fatalf("FromString(%q) = %+v, want an error", input, v)
			}
		})
	}
}

func TestString_ParsedVersion_RoundTrips(t *testing.T) {
	for _, input := range []string{"v1.2.3", "v1.2.3-rc1", "v0.0.10-0.20250101120000-abcdef123456"} {
		v, err := FromString(input)
		if err != nil {
			t.Fatalf("FromString(%q): %v", input, err)
		}
		if got := v.String(); got != input {
			t.Errorf("FromString(%q).String() = %q", input, got)
		}
	}
}

func TestCompare_OrdersByMajorMinorPatch(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"1.2.3", "1.2.3", 0},
		{"2.0.0", "1.9.9", 1},
		{"1.2.0", "1.10.0", -1},
		{"1.2.4", "1.2.3", 1},
		{"1.2.3-rc1", "1.2.3", 0}, // The suffix is not compared
	}

	for _, tt := range tests {
		a, _ := FromString(tt.a)
		b, _ := FromString(tt.b)
		if got := a.Compare(b); got != tt.want {
			t.Errorf("Compare(%s, %s) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
		if got := Compare(b, a); got != -tt.want {
			t.Errorf("Compare(%s, %s) = %d, want %d", tt.b, tt.a, got, -tt.want)
		}
	}
}
