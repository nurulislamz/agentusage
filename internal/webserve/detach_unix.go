//go:build unix

package webserve

import (
	"os"
	"strconv"
	"strings"
	"syscall"
)

func detachSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}

func ProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	if syscall.Kill(pid, 0) != nil {
		return false
	}
	return !processZombie(pid)
}

func processZombie(pid int) bool {
	data, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
	if err != nil {
		return false
	}
	s := string(data)
	i := strings.LastIndex(s, ")")
	if i < 0 || i+2 >= len(s) {
		return false
	}
	return s[i+2] == 'Z'
}

func terminatePID(pid int) error {
	return syscall.Kill(pid, syscall.SIGTERM)
}

func killPID(pid int) error {
	return syscall.Kill(pid, syscall.SIGKILL)
}
