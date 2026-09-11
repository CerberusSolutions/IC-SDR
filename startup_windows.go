//go:build windows

package main

import (
	"syscall"
	"unsafe"

	"go-zero/internal/i18n"
)

func showStartupError(message string) {
	user32 := syscall.NewLazyDLL("user32.dll")
	messageBox := user32.NewProc("MessageBoxW")
	text, _ := syscall.UTF16PtrFromString(message)
	title, _ := syscall.UTF16PtrFromString(i18n.T("IC-SDR - Error de arranque"))
	messageBox.Call(0, uintptr(unsafe.Pointer(text)), uintptr(unsafe.Pointer(title)), 0x10)
}
