package rtl433

import "math"

// frontend mirrors the original IC-SDR rtl_433 transport: it translates the
// selected channel, applies a 127-tap Blackman low-pass and decimates to CU8.
type frontend struct {
	inputRate, offsetHz  float64
	phase                float64
	taps                 []float64
	iRing, qRing         []float64
	position, decimation int
	outputRate           int
	gain                 float64
}

func newFrontend(inputRate, offsetHz float64, bandwidthHz int) *frontend {
	decimation := decimationForBandwidth(inputRate, bandwidthHz)
	f := &frontend{inputRate: inputRate, offsetHz: offsetHz, decimation: decimation, outputRate: int(math.Round(inputRate / float64(decimation))), gain: 1}
	if decimation > 1 {
		tapCount := 127
		if decimation == 2 {
			tapCount = 63
		} else if decimation == 4 {
			tapCount = 95
		}
		f.taps = lowPassTaps(tapCount, .45/float64(decimation))
		f.iRing, f.qRing = make([]float64, len(f.taps)), make([]float64, len(f.taps))
	}
	return f
}

func decimationForBandwidth(inputRate float64, bandwidthHz int) int {
	targetRate := 256_000
	if bandwidthHz >= 2_000_000 {
		targetRate = 2_048_000
	} else if bandwidthHz >= 1_000_000 {
		targetRate = 1_024_000
	} else if bandwidthHz >= 500_000 {
		targetRate = 512_000
	}
	return max(int(math.Round(inputRate/float64(targetRate))), 1)
}

func lowPassTaps(count int, cutoff float64) []float64 {
	taps, sum := make([]float64, count), float64(0)
	mid := float64(count-1) / 2
	for n := range taps {
		x := float64(n) - mid
		sinc := 2 * cutoff
		if x != 0 {
			sinc = math.Sin(2*math.Pi*cutoff*x) / (math.Pi * x)
		}
		window := .42 - .5*math.Cos(2*math.Pi*float64(n)/float64(count-1)) + .08*math.Cos(4*math.Pi*float64(n)/float64(count-1))
		taps[n] = sinc * window
		sum += taps[n]
	}
	for n := range taps {
		taps[n] /= sum
	}
	return taps
}

func (f *frontend) reset(offsetHz float64) {
	f.offsetHz, f.phase, f.position, f.gain = offsetHz, 0, 0, 1
	clear(f.iRing)
	clear(f.qRing)
}

func (f *frontend) process(iq []float32) []byte {
	if f.inputRate <= 0 {
		return nil
	}
	step := 2 * math.Pi * f.offsetHz / f.inputRate
	out := make([]byte, 0, len(iq)/f.decimation)
	peak := float64(0)
	for n := 0; n+1 < len(iq); n += 2 {
		i, q := float64(iq[n]), float64(iq[n+1])
		if step != 0 {
			c, s := math.Cos(f.phase), math.Sin(f.phase)
			i, q = i*c+q*s, q*c-i*s
			f.phase += step
			if f.phase > math.Pi {
				f.phase -= 2 * math.Pi
			} else if f.phase < -math.Pi {
				f.phase += 2 * math.Pi
			}
		}
		if f.decimation == 1 {
			peak = max(peak, math.Abs(i), math.Abs(q))
			out = append(out, cu8(i*f.gain), cu8(q*f.gain))
			continue
		}
		f.iRing[f.position], f.qRing[f.position] = i, q
		f.position = (f.position + 1) % len(f.taps)
		if (n/2)%f.decimation != f.decimation-1 {
			continue
		}
		fi, fq, index := float64(0), float64(0), f.position
		for k, tap := range f.taps {
			index--
			if index < 0 {
				index = len(f.taps) - 1
			}
			fi += f.iRing[index] * tap
			fq += f.qRing[index] * tap
			_ = k
		}
		peak = max(peak, math.Abs(fi), math.Abs(fq))
		out = append(out, cu8(fi*f.gain), cu8(fq*f.gain))
	}
	if peak > 0 {
		target := min(max(.88/peak, 1), 24)
		if target < f.gain {
			f.gain = target
		} else {
			f.gain = min(f.gain+.1, target)
		}
	}
	return out
}

func cu8(value float64) byte {
	value = min(max(value, -.992), .992)
	return byte(math.Round((value + 1) * 127.5))
}
