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
)

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

	command := fmt.Sprintf(`"%s" %%1`, helperPath)
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
	return FocusActivation{
		Protocol:  DefaultFocusProtocol,
		Helper:    helper,
		Arguments: FocusActivationArguments(pid),
	}, nil
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
