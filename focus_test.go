package toast

import "testing"

func TestFocusActivationArguments(t *testing.T) {
	if got := FocusActivationArguments(1234); got != "anfocus:1234" {
		t.Fatalf("FocusActivationArguments() = %q, want %q", got, "anfocus:1234")
	}
	if got := FocusActivationArguments(1234, "customfocus"); got != "customfocus:1234" {
		t.Fatalf("FocusActivationArguments(custom) = %q, want %q", got, "customfocus:1234")
	}
}
