package i18n

import (
	"regexp"
	"strings"
	"testing"
)

func TestParseAcceptsSavedAndDetectedTags(t *testing.T) {
	for _, c := range []struct {
		value string
		want  Language
	}{
		{"EN", English}, {"en", English}, {"en-GB", English},
		{"ES", Spanish}, {"es", Spanish}, {"es-ES", Spanish},
		{"", Spanish}, {"zz", Spanish},
	} {
		if got := Parse(c.value); got != c.want {
			t.Errorf("Parse(%q) = %q, want %q", c.value, got, c.want)
		}
	}
}

func TestLanguageForLocale(t *testing.T) {
	for _, c := range []struct {
		locale string
		want   Language
	}{
		{"es-ES", Spanish}, {"es_MX.UTF-8", Spanish}, {"ca-ES", Spanish},
		{"gl-ES", Spanish}, {"eu-ES", Spanish},
		{"en-GB", English}, {"en_US.UTF-8", English}, {"fr-FR", English},
		{"", Spanish}, {"C", Spanish}, {"POSIX", Spanish},
	} {
		if got := languageForLocale(c.locale); got != c.want {
			t.Errorf("languageForLocale(%q) = %q, want %q", c.locale, got, c.want)
		}
	}
}

func TestSpanishLeavesTextUntouched(t *testing.T) {
	restore := Current()
	defer Set(restore)
	Set(Spanish)
	if got := T("ESCÁNER"); got != "ESCÁNER" {
		t.Fatalf("T() in Spanish = %q, want the original", got)
	}
}

func TestEnglishTranslatesAndIsIdempotent(t *testing.T) {
	restore := Current()
	defer Set(restore)
	Set(English)
	once := T("ESCÁNER")
	if once != "SCANNER" {
		t.Fatalf("T(\"ESCÁNER\") = %q, want \"SCANNER\"", once)
	}
	// Text is translated where it reaches the screen, and some strings are
	// translated earlier as well (before they are clipped or wrapped), so T
	// must be safe to apply more than once.
	if twice := T(once); twice != once {
		t.Fatalf("T() is not idempotent: %q then %q", once, twice)
	}
	if unknown := T("PMR-01"); unknown != "PMR-01" {
		t.Fatalf("T() rewrote text with no catalogue entry: %q", unknown)
	}
}

func TestTfTranslatesTheFormatBeforeSubstituting(t *testing.T) {
	restore := Current()
	defer Set(restore)
	Set(English)
	if got := Tf("%d MEMORIAS", 7); got != "7 MEMORIES" {
		t.Fatalf("Tf() = %q, want \"7 MEMORIES\"", got)
	}
}

func TestSetNotifiesListeners(t *testing.T) {
	restore := Current()
	defer Set(restore)
	Set(Spanish)
	var seen []Language
	OnChange(func(language Language) { seen = append(seen, language) })
	Set(English)
	Set(English) // No change, so no second notification.
	Set(Spanish)
	if len(seen) != 2 || seen[0] != English || seen[1] != Spanish {
		t.Fatalf("listener saw %v, want [EN ES]", seen)
	}
}

func TestNextCyclesThroughEveryLanguage(t *testing.T) {
	restore := Current()
	defer Set(restore)
	Set(Spanish)
	if got := Next(); got != English {
		t.Fatalf("Next() after Spanish = %q, want English", got)
	}
	Set(English)
	if got := Next(); got != Spanish {
		t.Fatalf("Next() after English = %q, want Spanish", got)
	}
}

func TestDialFrequencyUsesTheLanguageDigitGrouping(t *testing.T) {
	restore := Current()
	defer Set(restore)
	Set(Spanish)
	if got := FormatDialFrequency(446_018_750); got != "446.018.750" {
		t.Fatalf("Spanish dial = %q, want \"446.018.750\"", got)
	}
	Set(English)
	if got := FormatDialFrequency(446_018_750); got != "446,018,750" {
		t.Fatalf("English dial = %q, want \"446,018,750\"", got)
	}
	// The dial always shows nine digits, grouped MHz, kHz then Hz, so its
	// clickable decades never move.
	if got := FormatDialFrequency(1_000); got != "0,001,000" {
		t.Fatalf("English dial for 1 kHz = %q, want \"0,001,000\"", got)
	}
}

func TestGroupDigits(t *testing.T) {
	restore := Current()
	defer Set(restore)
	Set(English)
	for _, c := range []struct {
		value int64
		want  string
	}{{0, "0"}, {7, "7"}, {999, "999"}, {1000, "1,000"}, {1_234_567, "1,234,567"}, {-2500, "-2,500"}} {
		if got := GroupDigits(c.value); got != c.want {
			t.Errorf("GroupDigits(%d) = %q, want %q", c.value, got, c.want)
		}
	}
	Set(Spanish)
	if got := GroupDigits(1_234_567); got != "1.234.567" {
		t.Fatalf("Spanish GroupDigits = %q, want \"1.234.567\"", got)
	}
}

var formatVerb = regexp.MustCompile(`%[-+ #0]*[0-9]*(?:\.[0-9]+)?[a-zA-Z%]`)

func TestCatalogueKeepsFormatVerbs(t *testing.T) {
	for source, target := range englishCatalogue {
		from := strings.Join(formatVerb.FindAllString(source, -1), " ")
		to := strings.Join(formatVerb.FindAllString(target, -1), " ")
		if from != to {
			t.Errorf("format verbs differ for %q:\n  Spanish: %s\n  English: %s", source, from, to)
		}
	}
}

func TestCatalogueEntriesAreUsable(t *testing.T) {
	for source, target := range englishCatalogue {
		if source == "" || target == "" {
			t.Errorf("empty catalogue entry: %q -> %q", source, target)
			continue
		}
		// Leading and trailing spaces carry meaning: they are how a prefix
		// joins the value that follows it. They have to survive translation.
		if strings.HasPrefix(source, " ") != strings.HasPrefix(target, " ") {
			t.Errorf("leading space lost or added: %q -> %q", source, target)
		}
		if strings.HasSuffix(source, " ") != strings.HasSuffix(target, " ") {
			t.Errorf("trailing space lost or added: %q -> %q", source, target)
		}
	}
}

func TestCatalogueUsesBritishSpelling(t *testing.T) {
	// A short list of the American spellings most likely to slip in.
	american := []string{"color", "center", "centered", "neighbor", "analyze", "organize", "license"}
	for source, target := range englishCatalogue {
		lower := strings.ToLower(target)
		for _, word := range american {
			if strings.Contains(lower, word) {
				t.Errorf("American spelling %q in translation of %q: %q", word, source, target)
			}
		}
	}
}
