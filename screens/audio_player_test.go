package screens

import (
	"math"
	"testing"
)

func TestDMRUsesBurstSizedPrebuffer(t *testing.T) {
	if got := playbackPrebufferSamples(true); got != audioSampleRate*340/1000 {
		t.Fatalf("digital prebuffer=%d", got)
	}
	if playbackPrebufferSamples(false) != analogPrebufferSamples {
		t.Fatal("analog prebuffer changed")
	}
}

func TestAudioPlayerConsumesLargestPostProcessedPeak(t *testing.T) {
	player := &AudioPlayer{}
	player.publishAudioPeak([]float32{.2, -.7, .4})
	player.publishAudioPeak([]float32{-.5})

	if got := player.ConsumeAudioPeak(); math.Abs(float64(got-.7)) > 1e-6 {
		t.Fatalf("ConsumeAudioPeak() = %v, want 0.7", got)
	}
	if got := player.ConsumeAudioPeak(); got != 0 {
		t.Fatalf("second ConsumeAudioPeak() = %v, want 0", got)
	}
}

func TestDialFrequencyAndDigitStep(t *testing.T) {
	if got := formatDialFrequency(446_018_750); got != "446.018.750" {
		t.Fatalf("formatDialFrequency() = %q", got)
	}
	screen := &MainScreen{tuningStepHz: 6_250, frequencyDigitExponent: 3}
	if got := screen.activeTuningStepHz(); got != 1_000 {
		t.Fatalf("activeTuningStepHz() = %d, want 1000", got)
	}
	screen.frequencyDigitExponent = -1
	if got := screen.activeTuningStepHz(); got != 6_250 {
		t.Fatalf("activeTuningStepHz() = %d, want 6250", got)
	}
}
