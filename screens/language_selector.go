package screens

import (
	rl "github.com/gen2brain/raylib-go/raylib"

	"go-zero/internal/i18n"
	"go-zero/simpleui"
)

// LanguageSelector is the two-flag control in the bottom settings strip. The
// active language is drawn at full strength with a highlighted border; the
// other is dimmed until the pointer is over it.
//
// It sits beside the theme button because both change how the interface looks
// rather than how the receiver behaves.
type LanguageSelector struct {
	simpleui.BaseElement
	languages []i18n.Language
	hovered   int
	pressed   int
	onChange  func(i18n.Language)
}

const (
	languageFlagWidth  = float32(58)
	languageFlagHeight = float32(34)
	languageFlagGap    = float32(12)
	languageFlagInset  = float32(8)
)

// LanguageSelectorWidth is the room the control needs, which the caller uses to
// place it in the settings strip.
const LanguageSelectorWidth = languageFlagInset*2 + languageFlagWidth*2 + languageFlagGap

// NewLanguageSelector builds the control. onChange is called with the newly
// selected language only when the selection actually changes.
func NewLanguageSelector(x, y float32, onChange func(i18n.Language)) *LanguageSelector {
	return &LanguageSelector{
		BaseElement: simpleui.NewBaseElement("language", x, y, LanguageSelectorWidth, 40),
		languages:   i18n.Languages,
		hovered:     -1,
		pressed:     -1,
		onChange:    onChange,
	}
}

// flagBounds returns the rectangle of the flag at index, centred vertically in
// the control.
func (s *LanguageSelector) flagBounds(index int) rl.Rectangle {
	bounds := s.Bounds()
	return rl.Rectangle{
		X:      bounds.X + languageFlagInset + float32(index)*(languageFlagWidth+languageFlagGap),
		Y:      bounds.Y + (bounds.Height-languageFlagHeight)/2,
		Width:  languageFlagWidth,
		Height: languageFlagHeight,
	}
}

func (s *LanguageSelector) Update(input simpleui.Input) bool {
	if !s.Enabled() {
		s.hovered, s.pressed = -1, -1
		return false
	}
	s.hovered = -1
	for index := range s.languages {
		if input.Over(s.flagBounds(index)) {
			s.hovered = index
			break
		}
	}
	if input.Pressed && s.hovered >= 0 {
		s.pressed = s.hovered
	}
	if input.Released {
		clicked := s.pressed >= 0 && s.pressed == s.hovered
		selected := s.pressed
		s.pressed = -1
		if clicked && s.languages[selected] != i18n.Current() {
			simpleui.PlayActivationFeedback()
			if s.onChange != nil {
				s.onChange(s.languages[selected])
			}
		}
	}
	return s.hovered >= 0 || s.pressed >= 0
}

// CapturingPointer keeps the press and the release on the same control.
func (s *LanguageSelector) CapturingPointer() bool { return s.pressed >= 0 }

func (s *LanguageSelector) Draw() {
	if !s.Visible() {
		return
	}
	active := i18n.Current()
	for index, language := range s.languages {
		bounds := s.flagBounds(index)
		selected := language == active
		drawLanguageFlag(language, bounds)
		if !selected {
			// Dim the language that is not in use, less so under the pointer,
			// so the control reads as a choice rather than as two decorations.
			veil := uint8(150)
			if s.hovered == index {
				veil = 70
			}
			rl.DrawRectangleRec(bounds, rl.Color{R: colors.background.R, G: colors.background.G, B: colors.background.B, A: veil})
		}
		border, thickness := colors.border, float32(1)
		if selected {
			border, thickness = colors.cyan, 2
		} else if s.hovered == index {
			border = colors.text
		}
		rl.DrawRectangleLinesEx(bounds, thickness, border)
	}
}

func drawLanguageFlag(language i18n.Language, bounds rl.Rectangle) {
	switch language {
	case i18n.English:
		drawUnionFlag(bounds)
	default:
		drawSpanishFlag(bounds)
	}
}
