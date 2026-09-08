//go:build !windows

package rtl433

import "syscall"

func hiddenProcessAttributes() *syscall.SysProcAttr { return nil }
