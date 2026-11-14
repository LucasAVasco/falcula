package status

import "testing"

func TestStatus_StringAndIsDoingNothing(t *testing.T) {
	tests := []struct {
		status    Status
		name      string
		doingNone bool
	}{
		{None, "None", true},
		{Preparing, "Preparing", false},
		{Ready, "Ready", true},
		{AbortingPrepare, "AbortingPrepare", false},
		{PrepareAborted, "PrepareAborted", true},
		{Running, "Running", false},
		{Ended, "Ended", true},
		{Stopping, "Stopping", false},
		{Stopped, "Stopped", true},
		{Disabled, "Disabled", false},
		{Error, "Error", false},
		{Status(100), "Unknown", false},
	}

	for _, tt := range tests {
		if got := tt.status.String(); got != tt.name {
			t.Errorf("Status(%d).String() = %q, want %q", tt.status, got, tt.name)
		}
		if got := tt.status.IsDoingNothing(); got != tt.doingNone {
			t.Errorf("%s.IsDoingNothing() = %v, want %v", tt.name, got, tt.doingNone)
		}
	}
}
