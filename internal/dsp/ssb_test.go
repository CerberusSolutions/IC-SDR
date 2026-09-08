package dsp

import (
	"math"
	"testing"
)

func TestSSBDemodulatorRecoversUSBAndLSB(t *testing.T) {
	for _, mode := range []string{"USB", "LSB"} {
		t.Run(mode, func(t *testing.T) {
			const inputRate, outputRate, offset, tone = 192000.0, 48000.0, 12000.0, 1000.0
			sign := 1.0
			if mode == "LSB" {
				sign = -1
			}
			iq := make([]float32, int(inputRate/2)*2)
			for sample := 0; sample < len(iq)/2; sample++ {
				phase := 2 * math.Pi * (offset + sign*tone) * float64(sample) / inputRate
				iq[sample*2], iq[sample*2+1] = float32(math.Cos(phase)), float32(math.Sin(phase))
			}
			demod := NewSSBDemodulator(inputRate, outputRate)
			audio := demod.Process(iq, mode, offset, 2400)
			if len(audio) < 23000 || len(audio) > 25000 {
				t.Fatalf("unexpected output length: %d", len(audio))
			}
			start := len(audio) / 2
			var sine, cosine float64
			for index, sample := range audio[start:] {
				phase := 2 * math.Pi * tone * float64(index) / outputRate
				sine += float64(sample) * math.Sin(phase)
				cosine += float64(sample) * math.Cos(phase)
			}
			if amplitude := 2 * math.Hypot(sine, cosine) / float64(len(audio)-start); amplitude < .02 {
				t.Fatalf("tone not recovered: %.4f", amplitude)
			}
		})
	}
}
