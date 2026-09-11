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
	all := append(scanned, computedGeometry()...)

	rl.SetTraceLogLevel(rl.LogError)
	rl.SetConfigFlags(rl.FlagWindowHidden)
	rl.InitWindow(320, 200, "uifit")
	defer rl.CloseWindow()

	overflows := check(all)
	byKind := map[string]int{}
	for _, m := range all {
		byKind[m.kind]++
	}
	fmt.Printf("uifit: %d texts checked in %d languages %v\n", len(all), len(i18n.Languages), byKind)
	if len(overflows) == 0 {
		fmt.Println("uifit: every string fits its control")
		return
	}
	for _, o := range overflows {
		fmt.Printf("uifit: OVERFLOW %-36s %-9s avail=%6.1f width=%6.1f  %q\n", o.where, o.kind, o.avail, o.width, o.drawn)
	}
	fmt.Printf("uifit: %d strings overflow\n", len(overflows))
	os.Exit(1)
}

type overflow struct {
	measurement
	drawn string
	width float32
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
		for _, language := range i18n.Languages {
			i18n.Set(language)
			drawn := composedTranslation(item.text)
			width := simpleui.MeasureTextStyledRaw(drawn, item.size, item.style).X
			if width > item.avail {
				found = append(found, overflow{item, drawn, width})
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
		return translated
	}
	for _, prefix := range []string{"ESTILO  ", "REANUDAR ", "FORMATO ", "FILTRO ", "BANDA ", "ESCUCHA "} {
		if len(text) > len(prefix) && text[:len(prefix)] == prefix {
			return i18n.T(prefix) + i18n.T(text[len(prefix):])
		}
	}
	return text
}
