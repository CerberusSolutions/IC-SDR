package tetra

import (
	"math"
	"testing"
)

func phasesFor(pattern []byte, errors []float64) []float64 {
	out := make([]float64, len(pattern)/2)
	for i := range out {
		magnitude := math.Pi / 4
		if pattern[i*2+1] != 0 {
			magnitude = 3 * math.Pi / 4
		}
		if pattern[i*2] != 0 {
			magnitude = -magnitude
		}
		out[i] = magnitude + errors[i%len(errors)]
	}
	return out
}

func TestDetectTrainingAcceptsDistortedSequences(t *testing.T) {
	errors := []float64{0.18, -0.22, 0.09, -0.14}
	for _, tc := range []struct {
		pattern []byte
		want    trainingKind
	}{
		{normalTraining, trainingNDB1},
		{normalTraining2, trainingNDB2},
		{syncTraining, trainingSync},
	} {
		got, metric := detectTraining(phasesFor(tc.pattern, errors))
		if got != tc.want {
			t.Fatalf("training=%v want=%v metric=%f", got, tc.want, metric)
		}
	}
}

func TestDetectTrainingRejectsUnrelatedPhase(t *testing.T) {
	phases := make([]float64, 32)
	for i := range phases {
		phases[i] = float64((i*7)%13)/13*2*math.Pi - math.Pi
	}
	if got, _ := detectTraining(phases); got != trainingNone {
		t.Fatalf("noise classified as %v", got)
	}
}

func TestTimingPhaseUsesHysteresis(t *testing.T) {
	if got := selectTimingPhase([4]float64{.20, .19, .22, .21}, 0); got != 0 {
		t.Fatalf("small timing improvement caused a jump to %d", got)
	}
	if got := selectTimingPhase([4]float64{.20, .11, .22, .21}, 0); got != 1 {
		t.Fatalf("clear timing improvement not selected: %d", got)
	}
}
