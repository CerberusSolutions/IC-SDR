package tetra

import "testing"

func TestViterbiRecoversRateQuarterCodeword(t *testing.T) {
	input := make([]byte, 80)
	for i := range input {
		input[i] = byte((i*7 + i/3) & 1)
	}
	poly := [4]byte{0x13, 0x1d, 0x17, 0x1b}
	parity := func(v byte) byte { v ^= v >> 4; v ^= v >> 2; v ^= v >> 1; return v & 1 }
	state := byte(0)
	coded := make([]byte, 0, 320)
	for _, bit := range input {
		sr := (state << 1) | bit
		for _, p := range poly {
			coded = append(coded, parity(sr&p))
		}
		state = sr & 15
	}
	got := viterbi(coded, 80)
	for i := range input {
		if got[i] != input[i] {
			t.Fatalf("bit %d=%d, want %d", i, got[i], input[i])
		}
	}
}

func TestTrainingRequiresCompleteExactSuffix(t *testing.T) {
	bits := append([]byte{0, 1, 0}, syncTraining...)
	if !suffixMatches(bits, syncTraining) {
		t.Fatal("exact SYNC sequence was not recognized")
	}
	bits[len(bits)-1] ^= 1
	if suffixMatches(bits, syncTraining) {
		t.Fatal("corrupted SYNC sequence was accepted")
	}
	if _, ok := decodeBSCH(make([]byte, 119)); ok {
		t.Fatal("short BSCH accepted")
	}
}
