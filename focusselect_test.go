package toast

import (
	"strings"
	"testing"
)

func TestSelectHostWindow_CurrentSemantics(t *testing.T) {
	// 内层祖先无可用窗口（无标题/有 owner/不可见），外层祖先命中。
	chain := []AncestorWindows{
		{PID: 100, Exe: "pwsh.exe", Windows: nil},
		{PID: 101, Exe: "OpenConsole.exe", Windows: []WindowInfo{
			{HWND: 0xAAA, Title: "", Visible: true, HasOwner: false},               // 无标题→不可用
			{HWND: 0xBBB, Title: "dlg", Visible: true, HasOwner: true},             // 有 owner→不可用
			{HWND: 0xCCC, Title: "hidden", Visible: false, HasOwner: false},        // 不可见→不可用
		}},
		{PID: 102, Exe: "WindowsTerminal.exe", Windows: []WindowInfo{
			{HWND: 0xD00, Title: "pwsh", Visible: true, HasOwner: false},           // ✓ 命中
			{HWND: 0xD01, Title: "other", Visible: true, HasOwner: false},
		}},
	}
	hwnd, pid, reason := SelectHostWindow(chain)
	if hwnd != 0xD00 || pid != 102 {
		t.Fatalf("got (hwnd=%#x pid=%d), want (0xD00, 102); reason=%q", hwnd, pid, reason)
	}
}

func TestSelectHostWindow_NoneUsable(t *testing.T) {
	chain := []AncestorWindows{
		{PID: 1, Exe: "a", Windows: []WindowInfo{{HWND: 1, Title: "", Visible: true}}},
	}
	hwnd, pid, reason := SelectHostWindow(chain)
	if hwnd != 0 || pid != 0 {
		t.Fatalf("expected none selected, got (%#x,%d)", hwnd, pid)
	}
	if reason == "" {
		t.Fatal("expected a non-empty reason when nothing selected")
	}
}

func TestFocusDiagString(t *testing.T) {
	d := FocusDiag{
		StartPID: 100,
		Chain: []AncestorWindows{
			{PID: 102, Exe: "WindowsTerminal.exe", Windows: []WindowInfo{
				{HWND: 0xD00, Class: "CASCADIA_HOSTING_WINDOW_CLASS", Title: "pwsh", Visible: true},
			}},
		},
		SelectedHWND: 0xD00, SelectedPID: 102, Reason: "first usable",
	}
	s := d.String()
	for _, want := range []string{"100", "WindowsTerminal.exe", "CASCADIA_HOSTING_WINDOW_CLASS", "d00", "pwsh"} {
		if !strings.Contains(strings.ToLower(s), strings.ToLower(want)) {
			t.Fatalf("FocusDiag.String() missing %q in:\n%s", want, s)
		}
	}
}
