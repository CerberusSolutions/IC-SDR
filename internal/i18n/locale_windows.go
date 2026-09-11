//go:build windows

package i18n

import (
	"syscall"
	"unsafe"
)

// detectSystemLanguage asks Windows for the user's preferred locale and maps
// it onto an IC-SDR language. Spanish locales keep the original interface;
// everything else starts in English. Any failure falls back to Spanish so the
// author's users see no change when detection is unavailable.
func detectSystemLanguage() Language {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getLocaleName := kernel32.NewProc("GetUserDefaultLocaleName")
	if err := getLocaleName.Find(); err != nil {
		return Spanish
	}
	// LOCALE_NAME_MAX_LENGTH is 85 UTF-16 code units.
	buffer := make([]uint16, 85)
	length, _, _ := getLocaleName.Call(uintptr(unsafe.Pointer(&buffer[0])), uintptr(len(buffer)))
	if length == 0 {
		return Spanish
	}
	name := syscall.UTF16ToString(buffer[:length])
	return languageForLocale(name)
}
