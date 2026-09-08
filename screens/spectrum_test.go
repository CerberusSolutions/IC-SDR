package screens

import "testing"

func TestSpectrumYMatchesOriginalDisplayRange(t *testing.T) {
	const y, height = float32(215), float32(235)
	screen := &MainScreen{spectrumMinimumDB: -37, spectrumMaximumDB: 0}
	if got := screen.spectrumY(0, y, height); got != y+5 {
		t.Fatalf("0 dB y = %v, want %v", got, y+5)
	}
	if got := screen.spectrumY(-37, y, height); got != y+height {
		t.Fatalf("-37 dB y = %v, want %v", got, y+height)
	}
	if got := screen.spectrumY(-100, y, height); got != y+height {
		t.Fatalf("below-range value was not clipped: y = %v", got)
	}
}

func TestInterpolatedSpectrumValueSmoothsBetweenBins(t *testing.T) {
	screen := &MainScreen{spanHz: 4}
	screen.stats.SampleRate = 4
	spectrum := []float32{0, 10, 20, 30}
	value, visible := screen.interpolatedSpectrumValue(spectrum, .375)
	if !visible || value != 15 {
		t.Fatalf("interpolated value = %v, visible=%v; want 15, true", value, visible)
	}
}
