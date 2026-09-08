package dmr

import (
	"math"
	"testing"
)

func TestFrontendProduces48kDiscriminator(t *testing.T) {
	const inputRate = 2_048_000
	front := newFrontend(inputRate, 0, 12_500)
	iq := make([]float32, inputRate/10*2)
	phase := 0.0
	for index := 0; index < len(iq); index += 2 {
		phase += 2 * math.Pi * 1000 / inputRate
		iq[index], iq[index+1] = float32(math.Cos(phase)), float32(math.Sin(phase))
	}
	pcm, measured := front.process(iq)
	if len(pcm) < 4700 || len(pcm) > 4900 {
		t.Fatalf("output samples=%d, want about 4800", len(pcm))
	}
	if measured < 850 || measured > 1150 {
		t.Fatalf("measured=%f Hz, want about 1000 Hz", measured)
	}
}

func TestFrontendIsContinuousAcrossIQBlocks(t *testing.T) {
	const inputRate = 2_048_000
	iq := make([]float32, inputRate/20*2)
	phase := 0.0
	for index := 0; index < len(iq); index += 2 {
		phase += 2 * math.Pi * 725 / inputRate
		iq[index], iq[index+1] = float32(math.Cos(phase)), float32(math.Sin(phase))
	}
	oneBlock := newFrontend(inputRate, 0, 12_500)
	want, _ := oneBlock.process(iq)
	chunked := newFrontend(inputRate, 0, 12_500)
	var got []int16
	for start := 0; start < len(iq); start += 4094 {
		end := min(start+4094, len(iq))
		part, _ := chunked.process(iq[start:end])
		got = append(got, part...)
	}
	if len(got) != len(want) {
		t.Fatalf("chunked samples=%d, one block=%d", len(got), len(want))
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("sample[%d]=%d, want %d", index, got[index], want[index])
		}
	}
}
