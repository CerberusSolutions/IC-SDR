// Package i18n provides the runtime language layer for IC-SDR.
//
// The original interface is written in Spanish and those Spanish strings stay
// in place as the message identifiers, so the source keeps reading naturally
// for its author. Translation happens where text reaches the screen: SimpleUI
// is handed T as its translator, so every label, button and drawn string is
// looked up on its way out. Text assembled with fmt.Sprintf cannot be matched
// after formatting, so those call sites use Tf to translate the format string
// before the arguments are substituted.
//
// A message with no catalogue entry is returned unchanged. That makes T
// idempotent on English text and safe to apply more than once, which matters
// for strings that are translated early (before wrapping or clipping) and
// again when SimpleUI finally draws them.
package i18n

import (
	"fmt"
	"sync"
)

// Language identifies one of the interface languages IC-SDR ships with.
type Language string

const (
	// Spanish is the language the interface was originally written in.
	Spanish Language = "ES"
	// English is British English.
	English Language = "EN"
)

// Languages lists the selectable languages in toggle order.
var Languages = []Language{Spanish, English}

var (
	mu        sync.RWMutex
	current   = Spanish
	listeners []func(Language)
)

// Valid reports whether value names a language IC-SDR can display.
func Valid(value Language) bool {
	return value == Spanish || value == English
}

// Parse converts a persisted or detected language tag into a Language,
// falling back to Spanish when the value is empty or unrecognised.
func Parse(value string) Language {
	switch {
	case len(value) >= 2 && (value[:2] == "EN" || value[:2] == "en"):
		return English
	case len(value) >= 2 && (value[:2] == "ES" || value[:2] == "es"):
		return Spanish
	default:
		return Spanish
	}
}

// Current returns the language the interface is displaying.
func Current() Language {
	mu.RLock()
	defer mu.RUnlock()
	return current
}

// Set switches the interface language and notifies every registered listener.
// Setting the language already in use does nothing, so callers may set it
// unconditionally while restoring saved settings.
func Set(language Language) {
	if !Valid(language) {
		return
	}
	mu.Lock()
	if current == language {
		mu.Unlock()
		return
	}
	current = language
	notify := make([]func(Language), len(listeners))
	copy(notify, listeners)
	mu.Unlock()
	for _, listener := range notify {
		listener(language)
	}
}

// Next returns the language that follows the current one, so a single control
// can cycle through the available languages.
func Next() Language {
	active := Current()
	for index, language := range Languages {
		if language == active {
			return Languages[(index+1)%len(Languages)]
		}
	}
	return Languages[0]
}

// OnChange registers a listener invoked after the language changes. Listeners
// run on the goroutine that called Set, which is the interface goroutine.
func OnChange(listener func(Language)) {
	if listener == nil {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	listeners = append(listeners, listener)
}

// T translates a Spanish source string into the active language. Text with no
// catalogue entry is returned unchanged, so T is safe to apply repeatedly and
// harmless on strings that are already English.
func T(text string) string {
	if text == "" {
		return text
	}
	mu.RLock()
	language := current
	mu.RUnlock()
	if language == Spanish {
		return text
	}
	if translated, found := englishCatalogue[text]; found {
		return translated
	}
	return text
}

// Tf translates a Spanish format string and then applies the arguments. The
// format string is the catalogue key, so verbs and their order are preserved
// by the translation rather than by the formatted result.
func Tf(format string, args ...any) string {
	return fmt.Sprintf(T(format), args...)
}

// Catalogue exposes the Spanish-to-English message table for tests.
func Catalogue() map[string]string {
	copied := make(map[string]string, len(englishCatalogue))
	for source, translated := range englishCatalogue {
		copied[source] = translated
	}
	return copied
}

// Errorf translates a Spanish format string and returns the resulting error.
// It behaves exactly like fmt.Errorf, %w wrapping included.
//
// Errors built this way are shown to the user: the receiver puts its failure
// text straight into the status line. The message has to be translated where
// it is built, because once fmt has substituted the arguments the result no
// longer matches any catalogue key.
func Errorf(format string, args ...any) error {
	return fmt.Errorf(T(format), args...)
}
