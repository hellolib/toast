//go:build windows

// toast-focus sends a Windows toast notification whose click action focuses the
// terminal window that launched this process.
package main

import (
	"fmt"
	"os"

	"github.com/hellolib/toast"
)

func main() {
	parentPID := os.Getppid()
	activation, err := toast.PrepareFocusActivation(parentPID)
	if err != nil {
		fmt.Fprintln(os.Stderr, "prepare focus activation:", err)
		os.Exit(1)
	}

	fmt.Printf("helper: %s\nparent PID = %d\nlaunch = %s\n\n", activation.Helper, parentPID, activation.Arguments)
	fmt.Println("Click the notification to focus the terminal window.")

	if err := toast.Push("Click to focus the current terminal",
		toast.WithAppID("agent-notify"),
		toast.WithTitle("agent-notify"),
		toast.WithMessage(fmt.Sprintf("Click to focus terminal from parent PID %d", parentPID)),
		toast.WithActivationType("protocol"),
		toast.WithActivationArguments(activation.Arguments),
	); err != nil {
		fmt.Fprintln(os.Stderr, "send toast:", err)
		os.Exit(1)
	}

	fmt.Println("notification sent")
}
