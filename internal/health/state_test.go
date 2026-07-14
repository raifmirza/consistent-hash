package health

import "testing"

func TestState_String(t *testing.T) {
	tests := []struct {
		state    State
		expected string
	}{
		{StateHealthy, "Healthy"},
		{StateUnhealthy, "Unhealthy"},
		{StateUnknown, "Unknown"},
	}

	for _, tc := range tests {
		if got := tc.state.String(); got != tc.expected {
			t.Fatalf("expected %q, got %q", tc.expected, got)
		}
	}
}
