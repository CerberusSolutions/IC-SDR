//go:build !windows

package sstv

import "syscall"

func hiddenProcessAttributes() *syscall.SysProcAttr { return nil }
