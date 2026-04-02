//go:build windows

package process

import (
	"fmt"
	"os/exec"
)

func killProcessTree(pid int) error {
	cmd := exec.Command("taskkill", "/T", "/F", "/PID", fmt.Sprintf("%d", pid))
	return cmd.Run()
}
