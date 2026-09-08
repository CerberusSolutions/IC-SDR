//go:build !windows

package aircraft

import "syscall"

func hiddenProcessAttributes() *syscall.SysProcAttr { return nil }
