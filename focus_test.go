package toast

import (
	"encoding/base64"
	"testing"
)

func TestFocusActivationArguments(t *testing.T) {
	if got := FocusActivationArguments(1234); got != "anfocus:1234" {
		t.Fatalf("FocusActivationArguments() = %q, want %q", got, "anfocus:1234")
	}
	if got := FocusActivationArguments(1234, "customfocus"); got != "customfocus:1234" {
		t.Fatalf("FocusActivationArguments(custom) = %q, want %q", got, "customfocus:1234")
	}
}

func TestBuildFocusArguments(t *testing.T) {
	cases := []struct {
		name              string
		pid               int
		hwnd              uintptr
		logPath, protocol string
		want              string
	}{
		{"pid only", 1234, 0, "", "", "anfocus:1234"},
		{"pid+hwnd", 1234, 0x120a3e, "", "", "anfocus:1234:120a3e"},
		{"pid+hwnd+log", 1234, 0x120a3e, `C:\Users\x\.agent-notify\focus-helper.log`, "",
			"anfocus:1234:120a3e:" + base64.RawURLEncoding.EncodeToString([]byte(`C:\Users\x\.agent-notify\focus-helper.log`))},
		{"log without hwnd keeps slot", 1234, 0, "/tmp/a.log", "",
			"anfocus:1234:0:" + base64.RawURLEncoding.EncodeToString([]byte("/tmp/a.log"))},
		{"custom protocol", 7, 0x1f, "", "customfocus", "customfocus:7:1f"},
	}
	for _, c := range cases {
		if got := buildFocusArguments(c.pid, c.hwnd, c.logPath, c.protocol); got != c.want {
			t.Fatalf("%s: buildFocusArguments = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestParseFocusActivation(t *testing.T) {
	logp := `C:\a b\focus.log`
	uri := "anfocus:1234:120a3e:" + base64.RawURLEncoding.EncodeToString([]byte(logp))
	pid, hwnd, gotLog, err := ParseFocusActivation(uri)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if pid != 1234 || hwnd != 0x120a3e || gotLog != logp {
		t.Fatalf("got (%d, %#x, %q), want (1234, 0x120a3e, %q)", pid, hwnd, gotLog, logp)
	}

	// round-trip with build
	if got := buildFocusArguments(1234, 0x120a3e, logp, ""); got != uri {
		t.Fatalf("round-trip mismatch: %q vs %q", got, uri)
	}

	// legacy forms still parse
	p2, h2, l2, err := ParseFocusActivation("anfocus:99")
	if err != nil || p2 != 99 || h2 != 0 || l2 != "" {
		t.Fatalf("legacy pid-only parse wrong: (%d,%#x,%q,%v)", p2, h2, l2, err)
	}
	p3, h3, _, err := ParseFocusActivation("anfocus:99:1f")
	if err != nil || p3 != 99 || h3 != 0x1f {
		t.Fatalf("legacy pid+hwnd parse wrong: (%d,%#x,%v)", p3, h3, err)
	}

	// bad base64 → logPath empty, no error (degrade)
	if _, _, l, err := ParseFocusActivation("anfocus:1:2:!!!notb64!!!"); err != nil || l != "" {
		t.Fatalf("bad b64 should degrade to empty logpath, got (%q,%v)", l, err)
	}
}

