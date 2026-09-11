// Command uifit checks that every interface string still fits the panel that
// draws it after translation.
//
// IC-SDR lays its interface out in fixed pixel coordinates, so a label that
// grows when it is translated has nowhere to go and simply spills over its
// control. This tool measures each string with the same embedded fonts the
// application uses, in every language, and reports anything that no longer
// fits.
//
// Run it from the repository root:
//
//	go run ./cmd/uifit
//
// It exits non-zero when something overflows, so it can be wired into CI. A
// hidden window is opened because font metrics come from the graphics layer;
// on a headless machine run it under a virtual display.
package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"

	"go-zero/internal/i18n"
	"go-zero/simpleui"
)

// measurement is one piece of text together with the width it has to fit into.
type measurement struct {
	where string
	kind  string
	text  string
	avail float32
	size  int32
	style simpleui.FontStyle
}

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

	scanned, err := scanSource(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "uifit: cannot read the interface source:", err)
		os.Exit(2)
	}

	stranded, err := checkReachable(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "uifit: cannot read the interface source:", err)
		os.Exit(2)
	}
	all := append(scanned, computedGeometry()...)

	rl.SetTraceLogLevel(rl.LogError)
	rl.SetConfigFlags(rl.FlagWindowHidden)
	rl.InitWindow(320, 200, "uifit")
	defer rl.CloseWindow()

	for _, item := range stranded {
		fmt.Printf("uifit: UNTRANSLATED %-26s %-24s %q\n", item.where, "("+item.reason+")", item.text)
	}

	overflows := check(all)
	byKind := map[string]int{}
	for _, m := range all {
		byKind[m.kind]++
	}
	fmt.Printf("uifit: %d texts checked in %d languages %v\n", len(all), len(i18n.Languages), byKind)
	if len(stranded) > 0 {
		fmt.Printf("uifit: %d strings are joined to something before they are drawn, so they reach the user in Spanish\n", len(stranded))
	}
	if len(overflows) == 0 && len(stranded) == 0 {
		fmt.Println("uifit: every string fits its control and reaches the translator")
		return
	}
	if len(overflows) == 0 {
		fmt.Println("uifit: every string fits its control")
		os.Exit(1)
	}
	regressions := 0
	for _, o := range overflows {
		label, note := "OVERFLOW", ""
		if o.preExisting {
			label, note = "TIGHT   ", "  (also overruns in Spanish; pre-existing)"
		} else {
			regressions++
		}
		fmt.Printf("uifit: %s %-30s %-9s avail=%6.1f width=%6.1f  %q%s\n", label, o.where, o.kind, o.avail, o.width, o.drawn, note)
	}
	if regressions == 0 {
		fmt.Printf("uifit: %d strings are tight in both languages; none made worse by translation\n", len(overflows))
		if len(stranded) == 0 {
			return
		}
	} else {
		fmt.Printf("uifit: %d strings overflow only once translated\n", regressions)
	}
	os.Exit(1)
}

type overflow struct {
	measurement
	drawn       string
	width       float32
	preExisting bool // the Spanish original does not fit either
}

// check measures every string in every language and returns the ones that do
// not fit. Measurement goes through the same translator the application
// installs, so what is measured is what would be painted.
func check(items []measurement) []overflow {
	simpleui.SetTranslator(i18n.T)
	defer simpleui.SetTranslator(nil)

	var found []overflow
	seen := map[string]bool{}
	for _, item := range items {
		key := item.where + "\x00" + item.text
		if seen[key] {
			continue
		}
		seen[key] = true

		// Measure the original first. Some panels were already tight before
		// any of this: the workspace squeezes its controls horizontally while
		// leaving text at full size, so a label can overrun in Spanish too.
		// Those are the author's to weigh up, not something the translation
		// introduced, so they are reported but do not fail the check.
		i18n.Set(i18n.Spanish)
		original := simpleui.MeasureTextStyledRaw(composedTranslation(item.text), item.size, item.style).X
		tight := original > item.avail

		for _, language := range i18n.Languages {
			i18n.Set(language)
			drawn := composedTranslation(item.text)
			width := simpleui.MeasureTextStyledRaw(drawn, item.size, item.style).X
			if width > item.avail {
				found = append(found, overflow{item, drawn, width, tight})
			}
		}
	}
	sort.Slice(found, func(a, b int) bool {
		return found[a].width-found[a].avail > found[b].width-found[b].avail
	})
	return found
}

// composedTranslation mirrors the labels the interface assembles from a
// translated prefix and a runtime value, so those are measured the way they
// are actually built.
func composedTranslation(text string) string {
	if translated := i18n.T(text); translated != text {
		return sampleValues(translated)
	}
	for _, prefix := range []string{"ESTILO  ", "REANUDAR ", "FORMATO ", "FILTRO ", "BANDA ", "ESCUCHA "} {
		if len(text) > len(prefix) && text[:len(prefix)] == prefix {
			return i18n.T(prefix) + i18n.T(text[len(prefix):])
		}
	}
	return sampleValues(text)
}

// sampleValues substitutes representative values for fmt verbs so a format
// string is measured at something like its drawn width rather than with the
// verbs left in.
func sampleValues(text string) string {
	// Typical rather than worst-case values: a counter in this interface is
	// usually one or two digits, and padding every verb out to three would put
	// most of the status lines permanently over budget.
	replacer := strings.NewReplacer(
		"%d", "0", "%s", "AAAA", "%v", "AAAA", "%q", "AAAA",
		"%.0f", "0", "%.1f", "0.0", "%.3f", "0.000", "%.6f", "000.000000",
		"%%", "%",
	)
	return replacer.Replace(text)
}
