//go:build windows

// toast-focus-helper is launched by the anfocus: protocol. Build it with
// -ldflags="-H windowsgui" so clicking the toast does not flash a console.
package main

import (
	"os"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

var (
	kernel32              = syscall.NewLazyDLL("kernel32.dll")
	user32                = syscall.NewLazyDLL("user32.dll")
	procCreateToolhelp32  = kernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32First    = kernel32.NewProc("Process32FirstW")
	procProcess32Next     = kernel32.NewProc("Process32NextW")
	procCloseHandle       = kernel32.NewProc("CloseHandle")
	procEnumWindows       = user32.NewProc("EnumWindows")
	procGetWindowThreadPr = user32.NewProc("GetWindowThreadProcessId")
	procIsWindowVisible   = user32.NewProc("IsWindowVisible")
	procSetForegroundWnd  = user32.NewProc("SetForegroundWindow")
	procAllowSetFG        = user32.NewProc("AllowSetForegroundWindow")
	procShowWindowAsync   = user32.NewProc("ShowWindowAsync")
	procAttachThreadInp   = user32.NewProc("AttachThreadInput")
	procGetForegWnd       = user32.NewProc("GetForegroundWindow")
	procGetCurrentThID    = kernel32.NewProc("GetCurrentThreadId")
)

const (
	th32csSnapProcess = 0x00000002
	maxPath           = 260
	swRestore         = 9
	asfwAny           = ^uint32(0)
)

type processEntry32 struct {
	Size            uint32
	Usage           uint32
	ProcessID       uint32
	DefaultHeapID   uintptr
	ModuleID        uint32
	Threads         uint32
	ParentProcessID uint32
	PriClassBase    int32
	Flags           uint32
	ExeFile         [maxPath]uint16
}

func main() {
	if len(os.Args) < 2 {
		os.Exit(1)
	}

	pid, err := parseActivationPID(os.Args[1])
	if err != nil {
		os.Exit(1)
	}

	windowPID := findWindowPID(uint32(pid))
	if windowPID == 0 {
		os.Exit(1)
	}
	if foundHwnd == 0 {
		hasVisibleWindow(windowPID)
	}
	if foundHwnd == 0 {
		os.Exit(1)
	}

	procShowWindowAsync.Call(foundHwnd, uintptr(swRestore))
	setForeground(foundHwnd)
}

func parseActivationPID(uri string) (int, error) {
	s := uri
	if idx := strings.Index(s, ":"); idx >= 0 {
		s = s[idx+1:]
	}
	s = strings.TrimPrefix(s, "//")
	return strconv.Atoi(s)
}

func parentPID(pid uint32) uint32 {
	snap, _, _ := procCreateToolhelp32.Call(th32csSnapProcess, 0)
	if snap == 0 {
		return 0
	}
	defer procCloseHandle.Call(snap)

	entry := processEntry32{Size: uint32(unsafe.Sizeof(processEntry32{}))}
	ok, _, _ := procProcess32First.Call(snap, uintptr(unsafe.Pointer(&entry)))
	for ok != 0 {
		if entry.ProcessID == pid {
			return entry.ParentProcessID
		}
		ok, _, _ = procProcess32Next.Call(snap, uintptr(unsafe.Pointer(&entry)))
	}
	return 0
}

func findWindowPID(start uint32) uint32 {
	for current := start; current != 0; {
		if hasVisibleWindow(current) {
			return current
		}
		parent := parentPID(current)
		if parent == 0 || parent == current {
			break
		}
		current = parent
	}
	return 0
}

var foundHwnd uintptr

func hasVisibleWindow(pid uint32) bool {
	foundHwnd = 0
	procEnumWindows.Call(
		syscall.NewCallback(func(hwnd uintptr, _ uintptr) uintptr {
			var windowPID uint32
			procGetWindowThreadPr.Call(hwnd, uintptr(unsafe.Pointer(&windowPID)))
			visible, _, _ := procIsWindowVisible.Call(hwnd)
			if windowPID == pid && visible != 0 {
				foundHwnd = hwnd
				return 0
			}
			return 1
		}),
		0,
	)
	return foundHwnd != 0
}

func setForeground(hwnd uintptr) {
	procAllowSetFG.Call(uintptr(asfwAny))

	foreground, _, _ := procGetForegWnd.Call()
	if foreground != 0 {
		foregroundThread, _, _ := procGetWindowThreadPr.Call(foreground, 0)
		currentThread, _, _ := procGetCurrentThID.Call()
		if foregroundThread != currentThread {
			procAttachThreadInp.Call(foregroundThread, currentThread, 1)
			procSetForegroundWnd.Call(hwnd)
			procAttachThreadInp.Call(foregroundThread, currentThread, 0)
			return
		}
	}

	procSetForegroundWnd.Call(hwnd)
}
