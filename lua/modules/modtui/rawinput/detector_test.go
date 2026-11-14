package rawinput

import "testing"

func TestAltTDetector(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"alt+t", "\x1bt", true},
		{"alt+shift+t", "\x1bT", true},
		{"after other keys", "abc\x1bt", true},
		{"repeated escape", "\x1b\x1bt", true},
		{"plain t", "t", false},
		{"other alt key", "\x1bx", false},
		{"escape sequence that ends with t", "\x1b[1t", false},
		{"t after an interrupted escape", "\x1bxt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d altTDetector
			got := false
			for _, b := range []byte(tt.input) {
				got = got || d.Feed(b)
			}
			if got != tt.want {
				t.Fatalf("Feed(%q) detected = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestAltTDetector_SplitAcrossReads(t *testing.T) {
	var d altTDetector
	if d.Feed(escape) {
		t.Fatal("escape alone is not Alt+T")
	}
	if !d.Feed('t') {
		t.Fatal("the escape of a previous read must be remembered")
	}
	if d.Feed('t') {
		t.Fatal("the detector must reset after a detection")
	}
}
