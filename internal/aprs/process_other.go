//go:build !windows

package aprs

import "syscall"

func hiddenProcessAttributes() *syscall.SysProcAttr { return nil }
