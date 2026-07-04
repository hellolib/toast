//go:build windows

package toast

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"unsafe"
)

var (
	focusKernel32              = syscall.NewLazyDLL("kernel32.dll")
	focusUser32                = syscall.NewLazyDLL("user32.dll")
	focusProcCreateToolhelp32  = focusKernel32.NewProc("CreateToolhelp32Snapshot")
	focusProcProcess32First    = focusKernel32.NewProc("Process32FirstW")
	focusProcProcess32Next     = focusKernel32.NewProc("Process32NextW")
	focusProcCloseHandle       = focusKernel32.NewProc("CloseHandle")
	focusProcEnumWindows       = focusUser32.NewProc("EnumWindows")
	focusProcGetWindow         = focusUser32.NewProc("GetWindow")
	focusProcGetWindowTextLen  = focusUser32.NewProc("GetWindowTextLengthW")
	focusProcGetWindowThreadPr = focusUser32.NewProc("GetWindowThreadProcessId")
	focusProcIsWindowVisible   = focusUser32.NewProc("IsWindowVisible")
)

const (
	focusTH32CSSnapProcess = 0x00000002
	focusMaxPath           = 260
	focusGWOwner           = 4
)

type focusProcessEntry32 struct {
	Size            uint32
	Usage           uint32
	ProcessID       uint32
	DefaultHeapID   uintptr
	ModuleID        uint32
	Threads         uint32
	ParentProcessID uint32
	PriClassBase    int32
	Flags           uint32
	ExeFile         [focusMaxPath]uint16
}

// FindFocusHelper locates the helper executable used for toast click-to-focus.
// Explicit candidates are checked first, followed by conventional helper names
// next to the current executable.
func FindFocusHelper(candidates ...string) (string, error) {
	for _, candidate := range append(nonEmpty(candidates), defaultFocusHelperCandidates()...) {
		if candidate == "" {
			continue
		}
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}
	return "", errors.New("toast focus helper not found")
}

// RegisterFocusProtocol registers protocol as a user-local URL protocol that
// launches helperPath with the clicked toast URI.
func RegisterFocusProtocol(helperPath string, protocols ...string) error {
	if helperPath == "" {
		return errors.New("focus helper path is empty")
	}

	protocol := DefaultFocusProtocol
	if len(protocols) > 0 && protocols[0] != "" {
		protocol = protocols[0]
	}

	command := fmt.Sprintf(`"%s" "%%1"`, helperPath)
	commands := [][]string{
		{"add", `HKCU\Software\Classes\` + protocol, "/ve", "/d", "URL:toast focus", "/f"},
		{"add", `HKCU\Software\Classes\` + protocol, "/v", "URL Protocol", "/d", "", "/f"},
		{"add", `HKCU\Software\Classes\` + protocol + `\shell\open\command`, "/ve", "/d", command, "/f"},
	}

	for _, args := range commands {
		cmd := exec.Command("reg", args...)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("reg %s: %w: %s", args[1], err, string(out))
		}
	}
	return nil
}

// PrepareFocusActivation finds the helper, registers the default focus
// protocol, and returns the URI that should be passed to WithActivationArguments.
func PrepareFocusActivation(pid int, helperCandidates ...string) (FocusActivation, error) {
	helper, err := FindFocusHelper(helperCandidates...)
	if err != nil {
		return FocusActivation{}, err
	}
	if err := RegisterFocusProtocol(helper); err != nil {
		return FocusActivation{}, err
	}
	hwnd := findFocusWindow(uint32(pid))
	return FocusActivation{
		Protocol:  DefaultFocusProtocol,
		Helper:    helper,
		Arguments: focusActivationArguments(pid, hwnd, DefaultFocusProtocol),
		Window:    hwnd,
	}, nil
}

func findFocusWindow(start uint32) uintptr {
	for current := start; current != 0; {
		if hwnd := visibleTopLevelWindow(current); hwnd != 0 {
			return hwnd
		}
		parent := focusParentPID(current)
		if parent == 0 || parent == current {
			break
		}
		current = parent
	}
	return 0
}

func focusParentPID(pid uint32) uint32 {
	snap, _, _ := focusProcCreateToolhelp32.Call(focusTH32CSSnapProcess, 0)
	if snap == 0 {
		return 0
	}
	defer focusProcCloseHandle.Call(snap)

	entry := focusProcessEntry32{Size: uint32(unsafe.Sizeof(focusProcessEntry32{}))}
	ok, _, _ := focusProcProcess32First.Call(snap, uintptr(unsafe.Pointer(&entry)))
	for ok != 0 {
		if entry.ProcessID == pid {
			return entry.ParentProcessID
		}
		ok, _, _ = focusProcProcess32Next.Call(snap, uintptr(unsafe.Pointer(&entry)))
	}
	return 0
}

func visibleTopLevelWindow(pid uint32) uintptr {
	var found uintptr
	focusProcEnumWindows.Call(
		syscall.NewCallback(func(hwnd uintptr, _ uintptr) uintptr {
			var windowPID uint32
			focusProcGetWindowThreadPr.Call(hwnd, uintptr(unsafe.Pointer(&windowPID)))
			if windowPID != pid || !isUsableTopLevelWindow(hwnd) {
				return 1
			}
			found = hwnd
			return 0
		}),
		0,
	)
	return found
}

func isUsableTopLevelWindow(hwnd uintptr) bool {
	visible, _, _ := focusProcIsWindowVisible.Call(hwnd)
	if visible == 0 {
		return false
	}
	owner, _, _ := focusProcGetWindow.Call(hwnd, focusGWOwner)
	if owner != 0 {
		return false
	}
	titleLen, _, _ := focusProcGetWindowTextLen.Call(hwnd)
	return titleLen > 0
}

func defaultFocusHelperCandidates() []string {
	exe, err := os.Executable()
	if err != nil {
		return nil
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}

	dir := filepath.Dir(exe)
	base := filepath.Base(exe)
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)

	var candidates []string
	if suffix := strings.TrimPrefix(stem, "toast-focus"); suffix != stem && suffix != "" {
		candidates = append(candidates, filepath.Join(dir, "toast-focus-helper"+suffix+ext))
	}
	candidates = append(candidates,
		filepath.Join(dir, stem+"-focus-helper"+ext),
		filepath.Join(dir, stem+"-helper"+ext),
		filepath.Join(dir, defaultFocusHelperName),
	)
	if runtime.GOARCH == "arm64" {
		candidates = append([]string{filepath.Join(dir, "toast-focus-helper-arm64.exe")}, candidates...)
	}
	return candidates
}

func nonEmpty(items []string) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}
