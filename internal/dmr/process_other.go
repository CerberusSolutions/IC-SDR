//go:build !windows

package dmr

import "syscall"

func hiddenProcessAttributes() *syscall.SysProcAttr { return nil }
