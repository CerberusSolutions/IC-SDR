package aprs

import "math"

type frontend struct {
	inputRate, outputRate, offsetHz        float64
	bandwidth                              int
	phase, resample                        float64
	i1, q1, i2, q2, i3, q3                 float64
	previousI, previousQ, previousDisc, dc float64
	errorHz                                float32
}

func newFrontend(inputRate float64) *frontend {
	return &frontend{inputRate: inputRate, outputRate: 48000, bandwidth: 12500, previousI: 1}
}
func (f *frontend) reset(offset float64, bandwidth int) {
	*f = *newFrontend(f.inputRate)
	f.offsetHz = offset
	f.bandwidth = max(8500, bandwidth)
}
func (f *frontend) process(iq []float32) []int16 {
	step := 2 * math.Pi * f.offsetHz / f.inputRate
	cutoff := math.Min(float64(f.bandwidth)*.48, 14000)
	alpha := 1 - math.Exp(-2*math.Pi*cutoff/f.inputRate)
	dcR := math.Exp(-2 * math.Pi * 30 / f.outputRate)
	deviation := min(max(float64(f.bandwidth)/5, 1500), 5000)
	devRadians := 2 * math.Pi * deviation / f.inputRate
	errorAlpha := 1 - math.Exp(-1/(.35*f.inputRate))
	out := make([]int16, 0, int(float64(len(iq)/2)*f.outputRate/f.inputRate)+2)
	for n := 0; n+1 < len(iq); n += 2 {
		c, s := math.Cos(f.phase), math.Sin(f.phase)
		i := float64(iq[n])*c + float64(iq[n+1])*s
		q := float64(iq[n+1])*c - float64(iq[n])*s
		f.phase += step
		if f.phase > math.Pi {
			f.phase -= 2 * math.Pi
		} else if f.phase < -math.Pi {
			f.phase += 2 * math.Pi
		}
		f.i1 += alpha * (i - f.i1)
		f.q1 += alpha * (q - f.q1)
		f.i2 += alpha * (f.i1 - f.i2)
		f.q2 += alpha * (f.q1 - f.q2)
		f.i3 += alpha * (f.i2 - f.i3)
		f.q3 += alpha * (f.q2 - f.q3)
		cross := f.q3*f.previousI - f.i3*f.previousQ
		dot := f.i3*f.previousI + f.q3*f.previousQ
		angle := math.Atan2(cross, dot)
		f.previousI, f.previousQ = f.i3, f.q3
		instant := angle * f.inputRate / (2 * math.Pi)
		f.errorHz += float32(errorAlpha) * (float32(instant) - f.errorHz)
		f.resample += f.outputRate
		if f.resample >= f.inputRate {
			f.resample -= f.inputRate
			disc := angle / max(devRadians, 1e-6)
			f.dc = disc - f.previousDisc + dcR*f.dc
			f.previousDisc = disc
			value := int(math.Round(min(max(f.dc*.72, -1), 1) * 32767))
			out = append(out, int16(value))
		}
	}
	return out
}
