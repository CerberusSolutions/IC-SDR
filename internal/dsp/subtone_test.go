package dsp

import (
	"math"
	"testing"
)

func TestCTCSSDetectsStandardToneAndRejectsSilence(t *testing.T) {
	samples := make([]float32, 1800)
	for i := range samples {
		samples[i] = float32(.2 * math.Sin(2*math.Pi*123*float64(i)/6000))
	}
	tone, confidence, ok := detectCTCSS(samples)
	if !ok || math.Abs(tone-123) > .01 || confidence < .5 {
		t.Fatalf("tone=%v confidence=%v ok=%v", tone, confidence, ok)
	}
	clear(samples)
	if _, _, ok = detectCTCSS(samples); ok {
		t.Fatal("silence detected as CTCSS")
	}
}

func TestGolayCorrectsThreeErrors(t *testing.T) {
	// A valid parity-first word is found by encoding through the parity check;
	// search is tiny and keeps this test independent of an encoder.
	var valid uint32
	for word := uint32(0); word < 1<<23; word++ {
		if golaySyndrome(word) == 0 && ((word>>9)&7) == 4 && standardDCS(uint(word&0x1ff)) {
			valid = word
			break
		}
	}
	if valid == 0 {
		t.Fatal("no valid DCS Golay word found")
	}
	corrupted := valid ^ (1 << 2) ^ (1 << 9) ^ (1 << 20)
	corrected, errors, ok := golayCorrect(corrupted)
	if !ok || corrected != valid || errors != 3 {
		t.Fatalf("corrected=%x want=%x errors=%d ok=%v", corrected, valid, errors, ok)
	}
}
