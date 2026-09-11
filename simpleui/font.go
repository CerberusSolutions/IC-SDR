package simpleui

import (
	_ "embed"
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// FontStyle selects one of the typefaces embedded in SimpleUI.
type FontStyle uint8

const (
	FontRegular FontStyle = iota
	FontSemiBold
	FontMono
)

//go:embed assets/fonts/Inter-Regular.ttf
var interRegularData []byte

//go:embed assets/fonts/Inter-SemiBold.ttf
var interSemiBoldData []byte

//go:embed assets/fonts/JetBrainsMono-Regular.ttf
var jetBrainsMonoData []byte

type fontKey struct {
	style FontStyle
	size  int32
}

var fonts = make(map[fontKey]rl.Font)
var textScale = float32(1)
var horizontalDrawScale = float32(1)

// SetHorizontalDrawScale compensates text rendered inside a horizontally
// scaled drawing region. Measurements remain unchanged; only glyph geometry is
// corrected so compact panels do not make their text narrow or faint.
func SetHorizontalDrawScale(scale float32) {
	if scale <= 0 {
		scale = 1
	}
	horizontalDrawScale = scale
}

// SetTextScale scales every SimpleUI text measurement and drawing operation.
// Call it before Run so control layout and rendering use the same metrics.
func SetTextScale(scale float32) {
	ensureNotStarted("SetTextScale")
	if scale < .5 || scale > 2 {
		panic("simpleui: text scale must be between 0.5 and 2")
	}
	textScale = scale
}

// DrawText draws smooth embedded-font text in logical coordinates. The text is
// localised through the installed translator; use DrawTextRaw for content that
// belongs to the user rather than to the interface.
func DrawText(text string, x, y float32, size int32, color rl.Color) {
	DrawTextStyled(text, x, y, size, FontRegular, color)
}

func DrawTextStyled(text string, x, y float32, size int32, style FontStyle, color rl.Color) {
	drawTextStyled(translate(text), x, y, size, style, color)
}

// DrawTextRaw draws text exactly as given, bypassing the translator. Use it for
// values the user typed or that arrived off the air, so a name that happens to
// match an interface string is never rewritten.
func DrawTextRaw(text string, x, y float32, size int32, color rl.Color) {
	drawTextStyled(text, x, y, size, FontRegular, color)
}

// DrawTextStyledRaw is DrawTextRaw with an explicit typeface.
func DrawTextStyledRaw(text string, x, y float32, size int32, style FontStyle, color rl.Color) {
	drawTextStyled(text, x, y, size, style, color)
}

func drawTextStyled(text string, x, y float32, size int32, style FontStyle, color rl.Color) {
	effectiveSize := scaledTextSize(size)
	font := cachedFont(style, effectiveSize)
	if horizontalDrawScale != 1 {
		rl.PushMatrix()
		rl.Translatef(x, 0, 0)
		rl.Scalef(1/horizontalDrawScale, 1, 1)
		rl.Translatef(-x, 0, 0)
		defer rl.PopMatrix()
	}
	rl.DrawTextEx(font, text, rl.Vector2{X: x, Y: y}, float32(effectiveSize), textSpacing(effectiveSize), color)
}

// MeasureText returns the logical size of embedded-font text once it has been
// localised, so callers lay out the text that will actually be drawn.
func MeasureText(text string, size int32) rl.Vector2 {
	return MeasureTextStyled(text, size, FontRegular)
}

func MeasureTextStyled(text string, size int32, style FontStyle) rl.Vector2 {
	return measureTextStyled(translate(text), size, style)
}

// MeasureTextRaw measures text exactly as given, matching DrawTextRaw.
func MeasureTextRaw(text string, size int32) rl.Vector2 {
	return measureTextStyled(text, size, FontRegular)
}

// MeasureTextStyledRaw is MeasureTextRaw with an explicit typeface.
func MeasureTextStyledRaw(text string, size int32, style FontStyle) rl.Vector2 {
	return measureTextStyled(text, size, style)
}

func measureTextStyled(text string, size int32, style FontStyle) rl.Vector2 {
	effectiveSize := scaledTextSize(size)
	font := cachedFont(style, effectiveSize)
	return rl.MeasureTextEx(font, text, float32(effectiveSize), textSpacing(effectiveSize))
}

func scaledTextSize(size int32) int32 {
	return max(1, int32(math.Round(float64(float32(size)*textScale))))
}

func cachedFont(style FontStyle, size int32) rl.Font {
	if size <= 0 {
		panic("simpleui: font size must be greater than zero")
	}
	key := fontKey{style: style, size: size}
	if font, exists := fonts[key]; exists {
		return font
	}
	data := fontData(style)
	font := rl.LoadFontFromMemory(".ttf", data, size*2, supportedCodepoints())
	if font.Texture.ID == 0 {
		panic(fmt.Sprintf("simpleui: failed to load embedded font style %d", style))
	}
	rl.SetTextureFilter(font.Texture, rl.FilterBilinear)
	fonts[key] = font
	return font
}

func fontData(style FontStyle) []byte {
	switch style {
	case FontSemiBold:
		return interSemiBoldData
	case FontMono:
		return jetBrainsMonoData
	default:
		return interRegularData
	}
}

func supportedCodepoints() []rune {
	codepoints := make([]rune, 0, 256)
	for value := rune(32); value <= 255; value++ {
		codepoints = append(codepoints, value)
	}
	return append(codepoints, '€', '₽', 'Ω', 'Δ', 'π', '∞', '≈', '≠', '≤', '≥', '±', '×', '÷', '√', '→', '←', '↑', '↓', '—', '–', '…', '•', '✓')
}

func textSpacing(size int32) float32 {
	return max(0.2, float32(size)*0.015)
}

func unloadFonts() {
	for key, font := range fonts {
		rl.UnloadFont(font)
		delete(fonts, key)
	}
}
