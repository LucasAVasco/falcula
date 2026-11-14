package version

import "testing"

func TestStringRoundTrip(t *testing.T) {
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
