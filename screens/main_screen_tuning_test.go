package screens

import "testing"

func TestTuneFixedByStepsMovesVFOWithoutMovingCenter(t *testing.T) {
	screen := NewMainScreen(nil)
	screen.frequencyHz = 14_261_000
	screen.centerFrequencyHz = 14_261_000
	screen.spanHz = 200_000
	screen.tuningStepHz = 1_000

	if centerChanged := screen.tuneFixedBySteps(3); centerChanged {
		t.Fatal("the IQ center moved while the VFO was away from the edge")
	}
	if got, want := screen.frequencyHz, int64(14_264_000); got != want {
		t.Fatalf("frequency = %d, want %d", got, want)
	}
	if got, want := screen.centerFrequencyHz, int64(14_261_000); got != want {
		t.Fatalf("center = %d, want %d", got, want)
	}
}

func TestTuneFixedAtFractionMovesCursorWithoutMovingFFT(t *testing.T) {
	screen := &MainScreen{centerFrequencyHz: 100_000_000, frequencyHz: 100_000_000, spanHz: 200_000, tuningStepHz: 1_000}
	screen.tuneFixedAtFraction(.75)
	if screen.frequencyHz != 100_050_000 {
		t.Fatalf("frequency = %d, want 100050000", screen.frequencyHz)
	}
	if screen.centerFrequencyHz != 100_000_000 {
		t.Fatalf("FFT center moved to %d", screen.centerFrequencyHz)
	}
}

func TestTuneFixedAtFractionRoundsToNearestStep(t *testing.T) {
	screen := &MainScreen{centerFrequencyHz: 446_100_000, frequencyHz: 446_100_000, spanHz: 200_000, tuningStepHz: 6_250}
	// Raw click is 446,018,000 Hz; the nearest 6.25 kHz channel is 446,018,750.
	screen.tuneFixedAtFraction(.09)
	if screen.frequencyHz != 446_018_750 {
		t.Fatalf("frequency = %d, want 446018750", screen.frequencyHz)
	}
	if screen.frequencyHz%screen.tuningStepHz != 0 {
		t.Fatalf("frequency %d does not respect step %d", screen.frequencyHz, screen.tuningStepHz)
	}
	if screen.centerFrequencyHz != 446_100_000 {
		t.Fatalf("FFT center moved to %d", screen.centerFrequencyHz)
	}
}

func TestViewGeometriesAndToolSelection(t *testing.T) {
	screen := &MainScreen{viewMode: 2, activeTool: "FFT"}
	_, _, _, fftHeight := screen.spectrumGeometry()
	_, waterfallY, _, waterfallHeight := screen.waterfallGeometry()
	if fftHeight != 470 || waterfallY != 685 || waterfallHeight != 141 {
		t.Fatalf("view 2 geometry = FFT %.0f, waterfall y %.0f h %.0f", fftHeight, waterfallY, waterfallHeight)
	}
	screen.setViewMode(3)
	_, _, _, fftHeight = screen.spectrumGeometry()
	_, _, _, waterfallHeight = screen.waterfallGeometry()
	if fftHeight != 235 || waterfallHeight != 340 {
		t.Fatalf("view 3 geometry = FFT %.0f, waterfall %.0f", fftHeight, waterfallHeight)
	}
	screen.selectTool("PBT_AUDIO")
	if screen.viewMode != 1 || screen.activeTool != "PBT_AUDIO" {
		t.Fatalf("tool selection did not restore view 1: view=%d tool=%s", screen.viewMode, screen.activeTool)
	}
}

func TestTuneFixedByStepsPansOnlyAtSpectrumEdge(t *testing.T) {
	screen := NewMainScreen(nil)
	screen.frequencyHz = 14_261_000
	screen.centerFrequencyHz = 14_261_000
	screen.spanHz = 200_000
	screen.tuningStepHz = 1_000
	screen.demodBandwidthHz = 9_000

	if centerChanged := screen.tuneFixedBySteps(100); !centerChanged {
		t.Fatal("the IQ center did not pan when the VFO reached the edge")
	}
	if got, want := screen.frequencyHz, int64(14_361_000); got != want {
		t.Fatalf("frequency = %d, want %d", got, want)
	}
	if got, want := screen.centerFrequencyHz, int64(14_271_000); got != want {
		t.Fatalf("center = %d, want %d", got, want)
	}
}

func TestWheelStepsPreservesDirection(t *testing.T) {
	for input, want := range map[float32]int64{1: 1, -1: -1, 2: 2, -2: -2, 0: 0} {
		if got := wheelSteps(input); got != want {
			t.Fatalf("wheelSteps(%v) = %d, want %d", input, got, want)
		}
	}
}

func TestTuneCenteredByStepsMovesFFTAndKeepsCursorCentered(t *testing.T) {
	screen := NewMainScreen(nil)
	screen.frequencyHz = 446_093_750
	screen.centerFrequencyHz = 446_093_750
	screen.tuningStepHz = 6_250
	screen.centerMode = true

	screen.tuneCenteredBySteps(2)

	if got, want := screen.frequencyHz, int64(446_106_250); got != want {
		t.Fatalf("frequency = %d, want %d", got, want)
	}
	if got, want := screen.centerFrequencyHz, screen.frequencyHz; got != want {
		t.Fatalf("center = %d, want tuned frequency %d", got, want)
	}
}

func TestChangeSpanRecentersWithoutChangingTunedFrequencyOrMode(t *testing.T) {
	screen := NewMainScreen(nil)
	screen.frequencyHz = 446_093_750
	screen.centerFrequencyHz = 446_000_000
	screen.spanHz = 500_000
	screen.centerMode = false

	screen.changeSpan(-1)

	if got, want := screen.spanHz, int64(250_000); got != want {
		t.Fatalf("span = %d, want %d", got, want)
	}
	if got, want := screen.frequencyHz, int64(446_093_750); got != want {
		t.Fatalf("tuned frequency changed to %d, want %d", got, want)
	}
	if got, want := screen.centerFrequencyHz, screen.frequencyHz; got != want {
		t.Fatalf("center = %d, want tuned frequency %d", got, want)
	}
	if screen.centerMode {
		t.Fatal("changing span changed FIX mode to CENTER")
	}
}
