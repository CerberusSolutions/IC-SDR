package tetra

import "math"

type SystemInfo struct {
	Valid                                   bool
	MCC, MNC                                uint16
	ColourCode, Timeslot, Frame, Multiframe uint8
}

const erasure = byte(0xff)

func decodeBSCH(input []byte) (SystemInfo, bool) {
	if len(input) != 120 {
		return SystemInfo{}, false
	}
	work := append([]byte(nil), input...)
	lfsr := uint32(3)
	for i := range work {
		st := func(y uint) uint32 { return lfsr >> (32 - y) }
		bit := (st(32) ^ st(26) ^ st(23) ^ st(22) ^ st(16) ^ st(12) ^ st(11) ^ st(10) ^ st(8) ^ st(7) ^ st(5) ^ st(4) ^ st(2) ^ st(1)) & 1
		work[i] ^= byte(bit)
		lfsr = (lfsr >> 1) | (bit << 31)
	}
	deint := make([]byte, 120)
	for k := 1; k <= 120; k++ {
		j := 1 + (11*k)%120
		deint[k-1] = work[j-1]
	}
	depunct := make([]byte, 320)
	src := 0
	keep := [8]bool{true, true, false, false, true, false, false, false}
	for i := range depunct {
		if keep[i&7] {
			depunct[i] = deint[src]
			src++
		} else {
			depunct[i] = erasure
		}
	}
	decoded := viterbi(depunct, 80)
	if crc16Bits(decoded[:76]) != 0x1d0f {
		return SystemInfo{}, false
	}
	value := func(start, n int) uint16 {
		var v uint16
		for i := 0; i < n; i++ {
			v = (v << 1) | uint16(decoded[start+i])
		}
		return v
	}
	return SystemInfo{Valid: true, MCC: value(31, 10), MNC: value(41, 14), ColourCode: uint8(value(4, 6)), Timeslot: uint8(value(10, 2) + 1), Frame: uint8(value(12, 5)), Multiframe: uint8(value(17, 6))}, true
}

func viterbi(coded []byte, outputBits int) []byte {
	var pm [16]int
	for i := 1; i < 16; i++ {
		pm[i] = 999
	}
	trace := make([][16]byte, outputBits)
	poly := [4]byte{0x13, 0x1d, 0x17, 0x1b}
	parity := func(v byte) byte { v ^= v >> 4; v ^= v >> 2; v ^= v >> 1; return v & 1 }
	for t := 0; t < outputBits; t++ {
		var next [16]int
		for i := range next {
			next[i] = 999
		}
		sym := coded[t*4 : t*4+4]
		for prev := 0; prev < 16; prev++ {
			if pm[prev] >= 999 {
				continue
			}
			for bit := byte(0); bit < 2; bit++ {
				ns := ((prev << 1) | int(bit)) & 15
				sr := byte(prev<<1) | bit
				m := pm[prev]
				for g := 0; g < 4; g++ {
					if sym[g] != erasure && parity(sr&poly[g]) != sym[g] {
						m++
					}
				}
				if m < next[ns] {
					next[ns] = m
					trace[t][ns] = byte(prev)
				}
			}
		}
		pm = next
	}
	best := 0
	for s := 1; s < 16; s++ {
		if pm[s] < pm[best] {
			best = s
		}
	}
	out := make([]byte, outputBits)
	for t := outputBits - 1; t >= 0; t-- {
		out[t] = byte(best & 1)
		best = int(trace[t][best])
	}
	return out
}
func crc16Bits(bits []byte) uint16 {
	crc := uint16(0xffff)
	for _, bit := range bits {
		b := uint16(bit) ^ ((crc >> 15) & 1)
		crc <<= 1
		if b != 0 {
			crc ^= 0x1021
		}
	}
	return crc
}

var syncTraining = []byte{1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 0, 1, 1, 1, 0, 0, 1, 1, 1, 0, 1, 0, 0, 1, 1, 1, 0, 0, 0, 0, 0, 1, 1, 0, 0, 1, 1, 1}
var normalTraining = []byte{1, 1, 0, 1, 0, 0, 0, 0, 1, 1, 1, 0, 1, 0, 0, 1, 1, 1, 0, 1, 0, 0}
var normalTraining2 = []byte{0, 1, 1, 1, 1, 0, 1, 0, 0, 1, 0, 0, 0, 0, 1, 1, 0, 1, 1, 1, 1, 0}

func suffixMatches(bits, pattern []byte) bool {
	if len(bits) < len(pattern) {
		return false
	}
	off := len(bits) - len(pattern)
	for i, b := range pattern {
		if bits[off+i] != b {
			return false
		}
	}
	return true
}

// trainingMetric compares differential phase samples with a TETRA training
// sequence. It retains the soft angular error instead of first reducing every
// sample to a hard dibit, so a noisy symbol does not necessarily lose a burst.
func trainingMetric(phases []float64, pattern []byte) float64 {
	symbols := len(pattern) / 2
	if symbols == 0 || len(phases) < symbols {
		return math.Inf(1)
	}
	start := len(phases) - symbols
	var sum float64
	for symbol := 0; symbol < symbols; symbol++ {
		magnitude := math.Pi / 4
		if pattern[symbol*2+1] != 0 {
			magnitude = 3 * math.Pi / 4
		}
		expected := magnitude
		if pattern[symbol*2] != 0 {
			expected = -expected
		}
		err := math.Atan2(math.Sin(phases[start+symbol]-expected), math.Cos(phases[start+symbol]-expected))
		sum += err * err
	}
	return sum / float64(symbols)
}

type trainingKind uint8

const (
	trainingNone trainingKind = iota
	trainingNDB1
	trainingNDB2
	trainingSync
)

func detectTraining(phases []float64) (trainingKind, float64) {
	metrics := [3]float64{
		trainingMetric(phases, normalTraining),
		trainingMetric(phases, normalTraining2),
		trainingMetric(phases, syncTraining),
	}
	best, second := 0, 1
	if metrics[second] < metrics[best] {
		best, second = second, best
	}
	for i := 2; i < len(metrics); i++ {
		if metrics[i] < metrics[best] {
			second, best = best, i
		} else if metrics[i] < metrics[second] {
			second = i
		}
	}
	// The absolute gate rejects noise. The margin gate avoids classifying an
	// ambiguous correlation shared by two training sequences.
	if metrics[best] > 0.42 || metrics[second]-metrics[best] < 0.06 {
		return trainingNone, metrics[best]
	}
	return trainingKind(best + 1), metrics[best]
}

func selectTimingPhase(scores [4]float64, current int) int {
	if current < 0 || current >= len(scores) {
		current = 0
	}
	best := current
	for phase := range scores {
		if scores[phase] < scores[best] {
			best = phase
		}
	}
	// Require a meaningful improvement. Close candidates otherwise cause a
	// one-sample oscillation that is worse than the small timing error itself.
	if best != current && scores[best] < scores[current]*0.88 {
		return best
	}
	return current
}
