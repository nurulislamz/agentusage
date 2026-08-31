//go:build windows

package webserve

import (
	"os"
	"syscall"
)

func detachSysProcAttr() *syscall.SysProcAttr {
	const (
		createNewProcessGroup = 0x00000200
		detachedProcess       = 0x00000008
	)
	return &syscall.SysProcAttr{
		CreationFlags: createNewProcessGroup | detachedProcess,
	}
}

func ProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	const processQueryLimitedInformation = 0x1000
	h, err := syscall.OpenProcess(processQueryLimitedInformation, false, uint32(pid))
	if err != nil {
		return false
	}
	_ = syscall.CloseHandle(h)
	return true
}

func terminatePID(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Kill()
}

func killPID(pid int) error {
	return terminatePID(pid)
}
