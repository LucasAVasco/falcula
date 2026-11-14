package colorgen

import "testing"

func TestNext_CyclesThroughThePalette(t *testing.T) {
	first := Next()
	seen := map[any]bool{first: true}
	for range len(colors) - 1 {
		seen[Next()] = true
	}

	if len(seen) != len(colors) {
		t.Fatalf("expected %d distinct colors in a cycle, got %d", len(colors), len(seen))
	}
	if Next() != first {
		t.Fatal("expected the palette to start again after a full cycle")
	}
}
