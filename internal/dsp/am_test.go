package dsp

import (
	"math"
	"testing"
)

func TestAMDemodulatorRecoversTone(t *testing.T) {
	const inputRate, outputRate, carrier, tone = 192000.0, 48000.0, 12000.0, 1000.0
	iq := make([]float32, int(inputRate/2)*2)
	for sample := 0; sample < len(iq)/2; sample++ {
		time := float64(sample) / inputRate
		envelope := 1 + .5*math.Sin(2*math.Pi*tone*time)
		iq[sample*2] = float32(envelope * math.Cos(2*math.Pi*carrier*time))
		iq[sample*2+1] = float32(envelope * math.Sin(2*math.Pi*carrier*time))
	}
	demod := NewAMDemodulator(inputRate, outputRate)
	audio := demod.Process(iq, carrier, 9000)
	if len(audio) < 23000 || len(audio) > 25000 {
		t.Fatalf("unexpected output length: %d", len(audio))
	}
	var energy float64
	for _, sample := range audio[len(audio)/2:] {
		energy += float64(sample * sample)
	}
	if energy/float64(len(audio)/2) < 1e-4 {
		t.Fatal("AM tone was not recovered")
	}
}
