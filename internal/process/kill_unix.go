//go:build !windows

package process

import (
	"syscall"
	"time"
)

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
