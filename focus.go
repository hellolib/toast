package toast

import "fmt"

const (
	// DefaultFocusProtocol is the protocol used by the bundled focus helper.
	DefaultFocusProtocol = "anfocus"

	defaultFocusHelperName = "toast-focus-helper.exe"
)

// FocusActivation describes the protocol activation data used to focus the
// source window when a Windows toast is clicked.
type FocusActivation struct {
	Protocol  string
	Helper    string
	Arguments string
	Window    uintptr
}

// FocusActivationArguments formats the URI passed to WithActivationArguments.
func FocusActivationArguments(pid int, protocols ...string) string {
	protocol := DefaultFocusProtocol
	if len(protocols) > 0 && protocols[0] != "" {
		protocol = protocols[0]
	}
	return fmt.Sprintf("%s:%d", protocol, pid)
}

func focusActivationArguments(pid int, hwnd uintptr, protocol string) string {
	if hwnd == 0 {
		return FocusActivationArguments(pid, protocol)
	}
	return fmt.Sprintf("%s:%d:%x", protocol, pid, hwnd)
}
