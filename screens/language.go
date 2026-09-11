package screens

import (
	"time"

	"go-zero/internal/i18n"
)

// setLanguage switches the interface language, refreshes the labels that were
// assembled once rather than rebuilt every frame, and remembers the choice.
//
// Nearly every string is translated on its way to the screen, so the change is
// visible immediately without rebuilding any control. Only labels that were
// concatenated when they were created need a nudge.
func (screen *MainScreen) setLanguage(language i18n.Language) {
	if language == i18n.Current() {
		return
	}
	i18n.Set(language)
	screen.language = language
	screen.refreshLocalisedLabels()
	screen.markSettingsDirty()
	// Write the choice out now rather than on the usual debounce. The map and
	// viewer windows are separate processes that learn the language by reading
	// this file, so a viewer opened in the next few hundred milliseconds would
	// otherwise start in the language the user has just left.
	screen.flushSettings(true)
}

// languageCheck throttles how often a viewer re-reads the settings file.
var languageCheck time.Time

// FollowPersistedLanguage keeps a viewer window in step with the main one.
// Viewers run as separate processes, so a language change in the receiver
// reaches them only through the settings file; re-reading it about once a
// second is cheap and needs no channel between the two.
func FollowPersistedLanguage() {
	if time.Since(languageCheck) < time.Second {
		return
	}
	languageCheck = time.Now()
	i18n.Set(StartupLanguage())
}

// refreshLocalisedLabels re-applies the labels that mix translated text with
// runtime values at the moment the control is built.
func (screen *MainScreen) refreshLocalisedLabels() {
	if screen.themeButton != nil {
		screen.themeButton.SetLabel(i18n.T("ESTILO  ") + i18n.T(themeDisplayName(screen.themeName)))
	}
	if screen.sdrSettings != nil {
		screen.sdrSettings.refresh()
	}
}
