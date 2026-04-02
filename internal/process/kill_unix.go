//go:build !windows

package process

import (
	"os/exec"
	"syscall"
	"time"
)

func hideWindow(cmd *exec.Cmd) {
	// No-op on Unix
}

func killProcessTree(pid int) error {
	// Send SIGTERM to process group
	if err := syscall.Kill(-pid, syscall.SIGTERM); err != nil {
		// Process may already be dead
		return nil
	}

	// Wait briefly, then force kill
	time.Sleep(2 * time.Second)
	syscall.Kill(-pid, syscall.SIGKILL)
	return nil
}
