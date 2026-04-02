package process

import (
	"runtime"
	"testing"
	"time"
)

func TestStartCommand_RunsProcess(t *testing.T) {
	mgr := NewManager("")

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "ping -n 10 127.0.0.1"
	} else {
		cmd = "sleep 10"
	}

	err := mgr.Start("test-cmd", cmd, t.TempDir())
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	status := mgr.Status("test-cmd")
	if status != Running {
		t.Errorf("expected Running, got %v", status)
	}

	// Cleanup
	mgr.KillAll()
}

func TestStartCommand_DetectsExit(t *testing.T) {
	mgr := NewManager("")

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "cmd /C echo hello"
	} else {
		cmd = "echo hello"
	}

	err := mgr.Start("test-cmd", cmd, t.TempDir())
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Wait for process to finish
	time.Sleep(500 * time.Millisecond)

	status := mgr.Status("test-cmd")
	if status != Stopped {
		t.Errorf("expected Stopped, got %v", status)
	}
}

func TestStartCommand_DetectsError(t *testing.T) {
	mgr := NewManager("")

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "cmd /C exit 1"
	} else {
		cmd = "sh -c 'exit 1'"
	}

	err := mgr.Start("test-cmd", cmd, t.TempDir())
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	status := mgr.Status("test-cmd")
	if status != Errored {
		t.Errorf("expected Errored, got %v", status)
	}
}

func TestKill_StopsRunningProcess(t *testing.T) {
	mgr := NewManager("")

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "ping -n 100 127.0.0.1"
	} else {
		cmd = "sleep 100"
	}

	mgr.Start("long-cmd", cmd, t.TempDir())

	time.Sleep(200 * time.Millisecond)
	if mgr.Status("long-cmd") != Running {
		t.Fatal("expected command to be running")
	}

	err := mgr.Kill("long-cmd")
	if err != nil {
		t.Fatalf("Kill failed: %v", err)
	}

	time.Sleep(500 * time.Millisecond)
	status := mgr.Status("long-cmd")
	if status == Running {
		t.Error("expected command to NOT be running after kill")
	}
}

func TestKillAll(t *testing.T) {
	mgr := NewManager("")

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "ping -n 100 127.0.0.1"
	} else {
		cmd = "sleep 100"
	}

	mgr.Start("cmd1", cmd, t.TempDir())
	mgr.Start("cmd2", cmd, t.TempDir())
	time.Sleep(200 * time.Millisecond)

	mgr.KillAll()
	time.Sleep(500 * time.Millisecond)

	if mgr.Status("cmd1") == Running {
		t.Error("cmd1 should not be running")
	}
	if mgr.Status("cmd2") == Running {
		t.Error("cmd2 should not be running")
	}
}
