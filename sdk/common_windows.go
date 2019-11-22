package sdk

import (
	"os"
	"syscall"
)

func Kill(pid int, signal syscall.Signal) error {
	p, e := os.FindProcess(pid)
	if e != nil {
		return e
	}
	return p.Signal(signal)
}
