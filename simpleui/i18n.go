package simpleui

// SimpleUI stays independent of any particular translation package. The host
// application installs a translator with SetTranslator and every string on its
// way to the screen passes through it, so labels created once at start-up
// follow a language change without being rebuilt.
//
// Measurement uses the same translator as drawing. Centring, right alignment
// and word wrapping all size the text that will actually be painted, so they
// stay correct in every language.

var translator func(string) string

// SetTranslator installs the function used to localise text before it is
// measured or drawn. Passing nil restores the untranslated behaviour.
//
// The translator is called on every drawn string, so it should be cheap and
// safe to call repeatedly with text it does not recognise.
func SetTranslator(translate func(string) string) {
	translator = translate
}

// translate localises text for display. It is deliberately a no-op when no
// translator is installed, which keeps SimpleUI usable on its own.
func translate(text string) string {
	if translator == nil || text == "" {
		return text
	}
	return translator(text)
}
