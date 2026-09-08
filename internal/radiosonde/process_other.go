//go:build !windows

package radiosonde

import "syscall"

func hiddenProcessAttributes() *syscall.SysProcAttr { return nil }
