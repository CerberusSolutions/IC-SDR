package screens

import (
	"testing"

	"go-zero/simpleui"
)

func TestValidTheme(t *testing.T) {
	for _, name := range []string{themeDark, themeLight, themeBlue, "dark", "Light"} {
		if !validTheme(name) {
			t.Fatalf("expected %q to be a valid theme", name)
		}
	}
	if validTheme("PURPLE") || validTheme("") {
		t.Fatal("unexpected theme accepted")
	}
}

func TestApplyThemeUpdatesPaletteAndControls(t *testing.T) {
	screen := &MainScreen{themeName: themeDark}
	t.Cleanup(func() { screen.applyTheme(themeDark) })

	screen.applyTheme(themeLight)
	if screen.themeName != themeLight || colors != lightPalette {
		t.Fatal("light palette was not applied")
	}
	if simpleui.CurrentTheme().Text != lightPalette.text {
		t.Fatal("SimpleUI did not receive the light palette")
	}

	screen.applyTheme(themeBlue)
	if screen.themeName != themeBlue || colors != bluePalette {
		t.Fatal("blue palette was not applied")
	}
}

func TestThemeCycleOrder(t *testing.T) {
	screen := &MainScreen{themeName: themeDark}
	t.Cleanup(func() { screen.applyTheme(themeDark) })

	screen.cycleTheme()
	if screen.themeName != themeLight {
		t.Fatalf("dark should advance to light, got %q", screen.themeName)
	}
	screen.cycleTheme()
	if screen.themeName != themeBlue {
		t.Fatalf("light should advance to blue, got %q", screen.themeName)
	}
	screen.cycleTheme()
	if screen.themeName != themeDark {
		t.Fatalf("blue should advance to dark, got %q", screen.themeName)
	}
}
