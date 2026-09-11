package i18n

import "strings"

// languageForLocale maps a platform locale name such as "es-ES", "en-GB" or
// "en_GB.UTF-8" onto an interface language. Spanish locales keep the original
// Spanish interface; every other locale starts in English.
func languageForLocale(locale string) Language {
	tag := strings.TrimSpace(locale)
	if tag == "" {
		return Spanish
	}
	if index := strings.IndexAny(tag, ".@"); index >= 0 {
		tag = tag[:index]
	}
	tag = strings.ReplaceAll(tag, "_", "-")
	primary := tag
	if index := strings.Index(tag, "-"); index >= 0 {
		primary = tag[:index]
	}
	switch strings.ToLower(primary) {
	case "es", "ca", "gl", "eu":
		// Spanish plus the co-official languages of Spain: a machine set to
		// Catalan, Galician or Basque is far more likely to want the Spanish
		// interface than the English one.
		return Spanish
	case "c", "posix", "":
		return Spanish
	default:
		return English
	}
}

// DetectSystemLanguage returns the language implied by the operating system's
// locale. It is used when no language has been saved yet.
func DetectSystemLanguage() Language {
	return detectSystemLanguage()
}
