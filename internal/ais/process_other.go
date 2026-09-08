//go:build !windows

package ais

import "syscall"

func hiddenProcessAttributes() *syscall.SysProcAttr { return nil }
