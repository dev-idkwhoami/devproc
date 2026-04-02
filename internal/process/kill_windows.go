//go:build windows

package process

import (
	"fmt"
	"os/exec"
	"syscall"
)

const CREATE_NO_WINDOW = 0x08000000

func hideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: CREATE_NO_WINDOW,
	}
}

func killProcessTree(pid int) error {
	cmd := exec.Command("taskkill", "/T", "/F", "/PID", fmt.Sprintf("%d", pid))
	hideWindow(cmd)
	return cmd.Run()
}
