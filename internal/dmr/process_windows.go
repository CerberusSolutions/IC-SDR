//go:build windows

package dmr

import "syscall"

// CREATE_NO_WINDOW prevents the console subsystem used by the bundled C++
// runner from flashing a command window whenever DMR is selected or resynced.
func hiddenProcessAttributes() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
}
