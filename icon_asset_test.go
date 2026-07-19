package toast

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileURI(t *testing.T) {
	cases := []struct{ in, want string }{
		{`C:\Users\x\.agent-notify\toast-icon.png`, "file:///C:/Users/x/.agent-notify/toast-icon.png"},
		{`C:\Program Files\a b\icon.png`, "file:///C:/Program%20Files/a%20b/icon.png"},
		{`C:\用户\icon.png`, "file:///C:/%E7%94%A8%E6%88%B7/icon.png"},
		{`/tmp/hellolib-toast/toast-icon.png`, "file:///tmp/hellolib-toast/toast-icon.png"},
	}
	for _, c := range cases {
		if got := fileURI(c.in); got != c.want {
			t.Fatalf("fileURI(%q)=%q want %q", c.in, got, c.want)
		}
	}
}

func TestWriteIconIfAbsent(t *testing.T) {
	p := filepath.Join(t.TempDir(), "sub", "icon.png")
	if err := writeIconIfAbsent(p, []byte("A")); err != nil {
		t.Fatal(err)
	}
	_ = writeIconIfAbsent(p, []byte("B")) // idempotent: must not overwrite
	got, _ := os.ReadFile(p)
	if string(got) != "A" {
		t.Fatalf("existing file overwritten: %q", got)
	}
}

func TestMaterializeDefaultIcon(t *testing.T) {
	t.Setenv("TMPDIR", t.TempDir()) // redirect os.TempDir() on darwin/linux
	p := materializeDefaultIcon()
	if p == "" || filepath.Base(p) != "toast-icon.png" {
		t.Fatalf("bad path %q", p)
	}
	data, err := os.ReadFile(p)
	if err != nil || len(data) != len(embeddedToastIcon) {
		t.Fatalf("icon not written correctly: len=%d err=%v", len(data), err)
	}
}

func TestEmbeddedToastIconIsPNG(t *testing.T) {
	if len(embeddedToastIcon) < 8 || string(embeddedToastIcon[1:4]) != "PNG" {
		t.Fatalf("embedded icon not a PNG (len=%d)", len(embeddedToastIcon))
	}
}
