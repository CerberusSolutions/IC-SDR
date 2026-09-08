package tetra

import (
	mathbits "math/bits"
	"sync"
)

// The AACH is carried in the two halves of the 30-bit broadcast block (BBK).
// It is protected by the shortened systematic RM(30,14) code from EN 300 392-2.
var (
	rm3014Once  sync.Once
	rm3014Words [1 << 14]uint32
)

var rm3014Generator = [14]uint16{
	0x9b60, 0x2de0, 0xfc20, 0xe03c, 0x983a, 0x5436, 0x2c2e,
	0xffdf, 0x8339, 0x42d5, 0x21ad, 0x1273, 0x0d6b, 0x09c7,
}

func initRM3014() {
	for message := 0; message < 1<<14; message++ {
		word := uint32(message) << 16
		for row := 0; row < 14; row++ {
			if message&(1<<(13-row)) != 0 {
				word ^= uint32(rm3014Generator[row])
			}
		}
		rm3014Words[message] = word
	}
}

func descrambleBlock(coded []byte, si SystemInfo) []byte {
	if !si.Valid {
		return nil
	}
	work := append([]byte(nil), coded...)
	lfsr := scramblingInit(si)
	for i := range work {
		st := func(y uint) uint32 { return lfsr >> (32 - y) }
		b := (st(32) ^ st(26) ^ st(23) ^ st(22) ^ st(16) ^ st(12) ^ st(11) ^ st(10) ^ st(8) ^ st(7) ^ st(5) ^ st(4) ^ st(2) ^ st(1)) & 1
		work[i] ^= byte(b)
		lfsr = (lfsr >> 1) | (b << 31)
	}
	return work
}

// decodeAACH returns the downlink usage marker. Values 0..3 are control or
// unallocated; values greater than 3 identify a traffic channel.
func decodeAACH(burst []byte, si SystemInfo) (uint8, bool) {
	if len(burst) < 282 || !si.Valid {
		return 0, false
	}
	bbk := make([]byte, 0, 30)
	bbk = append(bbk, burst[230:244]...)
	bbk = append(bbk, burst[266:282]...)
	decoded := descrambleBlock(bbk, si)
	var received uint32
	for _, bit := range decoded {
		received = received<<1 | uint32(bit&1)
	}
	rm3014Once.Do(initRM3014)
	bestMessage, bestDistance := 0, 31
	for message, word := range rm3014Words {
		distance := mathbits.OnesCount32(received ^ word)
		if distance < bestDistance {
			bestMessage, bestDistance = message, distance
			if distance == 0 {
				break
			}
		}
	}
	if bestDistance > 3 {
		return 0, false
	}
	// In frames other than frame 18, headers 1..3 carry DL usage in field 1.
	header := (bestMessage >> 12) & 3
	if header == 0 {
		return 0, true
	}
	return uint8((bestMessage >> 6) & 0x3f), true
}
