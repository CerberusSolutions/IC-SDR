package screens

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"go-zero/internal/i18n"
	"go-zero/simpleui"
)

// restoreLanguage puts the process language back after a test has changed it,
// so tests stay independent of the order they run in.
func restoreLanguage(t *testing.T) {
	t.Helper()
	previous := i18n.Current()
	t.Cleanup(func() { i18n.Set(previous) })
}

func TestLanguageSurvivesASettingsRoundTrip(t *testing.T) {
	restoreLanguage(t)
	i18n.Set(i18n.Spanish)

	path := filepath.Join(t.TempDir(), "settings.json")
	settings := persistedAppSettings{
		Version: appSettingsVersion, Language: string(i18n.English),
		BandCategory: "ISM", BandName: "PMR446", Mode: "NFM",
		FrequencyHz: 446_093_750, CenterFrequencyHz: 446_100_000,
		SpanHz: 500_000, TuningStepHz: 6_250,
	}
	if err := writeAppSettings(path, settings); err != nil {
		t.Fatal(err)
	}
	screen := NewMainScreen(nil)
	loadAppSettings(path, screen)
	if screen.language != i18n.English {
		t.Fatalf("restored language = %q, want %q", screen.language, i18n.English)
	}
	if i18n.Current() != i18n.English {
		t.Fatalf("loading settings did not apply the language: %q", i18n.Current())
	}
}

func TestStartupLanguageFallsBackToTheSystemLocale(t *testing.T) {
	restoreLanguage(t)
	// StartupLanguage reads the real settings file, which may or may not exist
	// on the machine running the tests. Either way it must return a language
	// the interface can display.
	if language := StartupLanguage(); !i18n.Valid(language) {
		t.Fatalf("StartupLanguage() = %q, which is not a usable language", language)
	}
}

func TestLegacySpanishAPRSViewIsMigrated(t *testing.T) {
	restoreLanguage(t)
	path := filepath.Join(t.TempDir(), "settings.json")
	// A settings file written by IC-SDR 0.3.1 or earlier: version 1, with the
	// APRS view stored under its Spanish name.
	legacy := persistedAppSettings{
		Version: appSettingsLegacyVersion, APRSView: "ESTACIONES",
		BandCategory: "ISM", BandName: "PMR446", Mode: "NFM",
		FrequencyHz: 446_093_750, CenterFrequencyHz: 446_100_000,
		SpanHz: 500_000, TuningStepHz: 6_250,
	}
	if err := writeAppSettings(path, legacy); err != nil {
		t.Fatal(err)
	}
	screen := NewMainScreen(nil)
	loadAppSettings(path, screen)
	if screen.aprsView != aprsViewStations {
		t.Fatalf("migrated APRS view = %q, want %q", screen.aprsView, aprsViewStations)
	}
	if screen.frequencyHz != legacy.FrequencyHz {
		t.Fatalf("version 1 settings were discarded instead of migrated")
	}
}

func TestUnknownAPRSViewIsIgnored(t *testing.T) {
	restoreLanguage(t)
	path := filepath.Join(t.TempDir(), "settings.json")
	settings := persistedAppSettings{
		Version: appSettingsVersion, APRSView: "NOT A VIEW",
		BandCategory: "ISM", BandName: "PMR446", Mode: "NFM",
		FrequencyHz: 446_093_750, CenterFrequencyHz: 446_100_000,
		SpanHz: 500_000, TuningStepHz: 6_250,
	}
	if err := writeAppSettings(path, settings); err != nil {
		t.Fatal(err)
	}
	screen := NewMainScreen(nil)
	loadAppSettings(path, screen)
	if screen.aprsView != aprsViewPackets {
		t.Fatalf("APRS view = %q, want the default %q", screen.aprsView, aprsViewPackets)
	}
}

func TestUngroupedMemoriesAreStoredWithoutTheSpanishSentinel(t *testing.T) {
	directory := t.TempDir()
	panel := &MemoryPanel{
		path: filepath.Join(directory, "memories.json"),
		memories: []MemoryEntry{
			{Name: "MEM 01", FrequencyHz: 446_006_250, Group: memoryGroupNone},
			{Name: "MEM 02", FrequencyHz: 446_018_750, Group: "PMR"},
		},
	}
	panel.save()

	data, err := os.ReadFile(panel.path)
	if err != nil {
		t.Fatal(err)
	}
	var stored []MemoryEntry
	if err := json.Unmarshal(data, &stored); err != nil {
		t.Fatal(err)
	}
	if stored[0].Group != "" {
		t.Fatalf("ungrouped memory stored group %q, want an empty group", stored[0].Group)
	}
	if stored[1].Group != "PMR" {
		t.Fatalf("named group was rewritten to %q", stored[1].Group)
	}
	// The in-memory list must be untouched: save writes a normalised copy.
	if panel.memories[0].Group != memoryGroupNone {
		t.Fatalf("save() mutated the live memory list: %q", panel.memories[0].Group)
	}
	// Reading either form back gives the sentinel again.
	if got := memoryGroup(stored[0]); got != memoryGroupNone {
		t.Fatalf("empty group read back as %q, want %q", got, memoryGroupNone)
	}
	if got := memoryGroup(MemoryEntry{Group: memoryGroupNone}); got != memoryGroupNone {
		t.Fatalf("legacy Spanish group read back as %q", got)
	}
}

func TestLanguageSelectorReportsTheActiveLanguage(t *testing.T) {
	restoreLanguage(t)
	i18n.Set(i18n.Spanish)

	var chosen i18n.Language
	selector := NewLanguageSelector(1000, 842, func(language i18n.Language) { chosen = language })
	if width := selector.Bounds().Width; width != LanguageSelectorWidth {
		t.Fatalf("selector width = %v, want %v", width, LanguageSelectorWidth)
	}
	// Both flags must sit inside the control and must not overlap.
	first, second := selector.flagBounds(0), selector.flagBounds(1)
	if first.X < selector.Bounds().X || second.X+second.Width > selector.Bounds().X+selector.Bounds().Width {
		t.Fatalf("flags fall outside the control: %+v %+v", first, second)
	}
	if first.X+first.Width >= second.X {
		t.Fatalf("flags overlap: %+v %+v", first, second)
	}
	if chosen != "" {
		t.Fatalf("building the selector reported a language change")
	}
}

func TestSetLanguageRefreshesComposedLabels(t *testing.T) {
	restoreLanguage(t)
	i18n.Set(i18n.Spanish)

	// The SimpleUI element manager is global and rejects duplicate IDs, so a
	// test cannot call CreateControls a second time. Build only the control
	// setLanguage has to refresh, exactly as CreateControls builds it.
	screen := NewMainScreen(nil)
	screen.themeButton = simpleui.NewButton("themeButtonUnderTest", 0, 0, 190, 40,
		i18n.T("ESTILO  ")+i18n.T(themeDisplayName(screen.themeName)), 13)
	before := screen.themeButton.Label()
	if before != "ESTILO  "+themeDisplayName(screen.themeName) {
		t.Fatalf("theme button label = %q, want the Spanish original", before)
	}
	screen.setLanguage(i18n.English)
	after := screen.themeButton.Label()
	if after == before {
		t.Fatalf("theme button label was not refreshed: %q", after)
	}
	if after != "THEME  "+i18n.T(themeDisplayName(screen.themeName)) {
		t.Fatalf("theme button label = %q, want the English form", after)
	}
	if screen.language != i18n.English || i18n.Current() != i18n.English {
		t.Fatalf("language not applied: screen=%q process=%q", screen.language, i18n.Current())
	}
	if !screen.settingsDirty {
		t.Fatalf("changing the language did not mark the settings for saving")
	}
}

func TestDialFrequencyFollowsTheLanguage(t *testing.T) {
	restoreLanguage(t)
	i18n.Set(i18n.Spanish)
	if got := formatDialFrequency(446_018_750); got != "446.018.750" {
		t.Fatalf("Spanish dial = %q", got)
	}
	i18n.Set(i18n.English)
	if got := formatDialFrequency(446_018_750); got != "446,018,750" {
		t.Fatalf("English dial = %q", got)
	}
	// Changing the separator must not move the digits the dial lets you click.
	if got := countFrequencyDigits(formatDialFrequency(446_018_750)); got != 9 {
		t.Fatalf("clickable digit count = %d, want 9", got)
	}
}
