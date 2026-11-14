package status

import "testing"

func TestStatus_StringAndIsDoingNothing(t *testing.T) {
	tests := []struct {
		status    Status
		name      string
		doingNone bool
		active    bool
	}{
		{None, "None", true, false},
		{Preparing, "Preparing", false, true},
		{Ready, "Ready", true, false},
		{AbortingPrepare, "AbortingPrepare", false, true},
		{PrepareAborted, "PrepareAborted", true, false},
		{Running, "Running", false, true},
		{Ended, "Ended", true, false},
		{Stopping, "Stopping", false, true},
		{Stopped, "Stopped", true, false},
		{Disabled, "Disabled", false, false},
		{Error, "Error", false, false},
		{Status(100), "Unknown", false, false},
	}

	for _, tt := range tests {
		if got := tt.status.String(); got != tt.name {
			t.Errorf("Status(%d).String() = %q, want %q", tt.status, got, tt.name)
		}
		if got := tt.status.IsActive(); got != tt.active {
			t.Errorf("%s.IsActive() = %v, want %v", tt.name, got, tt.active)
		}
		if got := tt.status.IsDoingNothing(); got != tt.doingNone {
			t.Errorf("%s.IsDoingNothing() = %v, want %v", tt.name, got, tt.doingNone)
		}
	}
}
