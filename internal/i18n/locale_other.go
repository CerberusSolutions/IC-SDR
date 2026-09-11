//go:build !windows

package i18n

import "os"

// detectSystemLanguage reads the usual POSIX locale variables. IC-SDR targets
// Windows, so this exists to keep development builds and tests on other
// platforms behaving the same way.
func detectSystemLanguage() Language {
	for _, variable := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if value := os.Getenv(variable); value != "" {
			return languageForLocale(value)
		}
	}
	return Spanish
}
