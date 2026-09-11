package screens

import "go-zero/internal/i18n"

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
