package dsp

import (
	"math"
	"testing"
)

func TestFFTPutsComplexToneAtExpectedCenteredBin(t *testing.T) {
	const size = 1024
	const toneBin = 123
	iq := make([]float32, size*2)
	for index := 0; index < size; index++ {
		phase := 2 * math.Pi * toneBin * float64(index) / size
		iq[index*2] = float32(math.Cos(phase))
		iq[index*2+1] = float32(math.Sin(phase))
	}
	spectrum := make([]float32, size)
	NewFFT(size).Process(iq, spectrum)
	maximum := 0
	for index := range spectrum {
		if spectrum[index] > spectrum[maximum] {
			maximum = index
		}
	}
	want := size/2 + toneBin
	if maximum != want {
		t.Fatalf("peak bin = %d, want %d", maximum, want)
	}
}

func TestFFTFullScaleComplexToneHasZeroDBFSPeak(t *testing.T) {
	const size = 1024
	const toneBin = 64
	iq := make([]float32, size*2)
	for index := 0; index < size; index++ {
		phase := 2 * math.Pi * toneBin * float64(index) / size
		iq[index*2] = float32(math.Cos(phase))
		iq[index*2+1] = float32(math.Sin(phase))
	}
	spectrum := make([]float32, size)
	NewFFT(size).Process(iq, spectrum)
	peak := spectrum[size/2+toneBin]
	if math.Abs(float64(peak)) > .02 {
		t.Fatalf("full-scale FFT peak = %.4f dBFS, want 0 dBFS", peak)
	}
}
