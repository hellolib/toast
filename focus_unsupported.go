//go:build !windows

package toast

import "errors"

// FindFocusHelper is only supported on Windows.
func FindFocusHelper(_ ...string) (string, error) {
	return "", errors.New("toast focus helper is only supported on windows")
}

// RegisterFocusProtocol is only supported on Windows.
func RegisterFocusProtocol(_ string, _ ...string) error {
	return errors.New("toast focus protocol is only supported on windows")
}

// PrepareFocusActivation is only supported on Windows.
func PrepareFocusActivation(_ int, _ ...string) (FocusActivation, error) {
	return FocusActivation{}, errors.New("toast focus activation is only supported on windows")
}
