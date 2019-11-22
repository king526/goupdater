//+build !windows

package sdk

import (
	"syscall"
)

func Kill(pid int, signal syscall.Signal) error {
	return syscall.Kill(pid, signal)
}
