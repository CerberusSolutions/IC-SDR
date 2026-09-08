package dmr

import (
	"math"
	"sync"
)

const (
	frontendFIRTaps   = 511
	frontendFIRPhases = 256
	frontendFIRHalf   = frontendFIRTaps / 2
	frontendHistory   = 4096
)

// frontend translates the selected RF channel, resamples it to the 48 kHz
// expected by DSDcc and produces a normalized FM discriminator stream. The
// polyphase FIR preserves the DMR 4FSK symbols and rejects adjacent channels.
type frontend struct {
	mu                                sync.Mutex
	inputRate, offsetHz, correctionHz float64
	bandwidth                         int
	ncoCos, ncoSin                    float64
	ncoSamples                        uint64
	iHistory, qHistory                [frontendHistory]float64
	inputSampleIndex                  int64
	nextOutputPosition                float64
	previousI, previousQ              float64
	kernel                            [][]float64
}

func newFrontend(rate, offset float64, bandwidth int) *frontend {
	f := &frontend{inputRate: rate}
	f.reset(offset, bandwidth)
	return f
}

func (f *frontend) reset(offset float64, bandwidth int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	bandwidth = min(max(bandwidth, 8_000), 18_000)
	if f.kernel == nil || bandwidth != f.bandwidth {
		f.kernel = buildResampleKernel(f.inputRate, math.Min(6500, float64(bandwidth)*.48))
	}
	f.offsetHz, f.correctionHz, f.bandwidth = offset, 0, bandwidth
	f.ncoCos, f.ncoSin, f.ncoSamples = 1, 0, 0
	f.inputSampleIndex, f.nextOutputPosition = 0, 0
	f.previousI, f.previousQ = 1, 0
	f.iHistory, f.qHistory = [frontendHistory]float64{}, [frontendHistory]float64{}
}

func buildResampleKernel(inputRate, cutoffHz float64) [][]float64 {
	result := make([][]float64, frontendFIRPhases)
	if inputRate <= 0 {
		return result
	}
	normalizedCutoff := cutoffHz / inputRate
	for phase := range result {
		fraction := float64(phase) / float64(frontendFIRPhases-1)
		result[phase] = make([]float64, frontendFIRTaps)
		sum := 0.0
		for tap := range result[phase] {
			distance := float64(tap-frontendFIRHalf) - fraction
			sinc := 2 * normalizedCutoff
			if math.Abs(distance) >= 1e-12 {
				sinc = math.Sin(2*math.Pi*normalizedCutoff*distance) / (math.Pi * distance)
			}
			window := .42 - .5*math.Cos(2*math.Pi*float64(tap)/float64(frontendFIRTaps-1)) +
				.08*math.Cos(4*math.Pi*float64(tap)/float64(frontendFIRTaps-1))
			result[phase][tap] = sinc * window
			sum += result[phase][tap]
		}
		if math.Abs(sum) > 1e-12 {
			for tap := range result[phase] {
				result[phase][tap] /= sum
			}
		}
	}
	return result
}

func (f *frontend) process(iq []float32) ([]int16, float32) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.inputRate <= 0 || len(f.kernel) != frontendFIRPhases {
		return nil, 0
	}
	phaseStep := 2 * math.Pi * (f.offsetHz + f.correctionHz) / f.inputRate
	stepCos, stepSin := math.Cos(phaseStep), math.Sin(phaseStep)
	inputPerOutput := f.inputRate / outputRate
	result := make([]int16, 0, int(float64(len(iq)/2)/inputPerOutput)+2)
	errorSum := 0.0
	for n := 0; n+1 < len(iq); n += 2 {
		i := float64(iq[n])*f.ncoCos + float64(iq[n+1])*f.ncoSin
		q := float64(iq[n+1])*f.ncoCos - float64(iq[n])*f.ncoSin
		oldCos := f.ncoCos
		f.ncoCos = oldCos*stepCos - f.ncoSin*stepSin
		f.ncoSin = f.ncoSin*stepCos + oldCos*stepSin
		f.ncoSamples++
		if f.ncoSamples%4096 == 0 {
			magnitude := math.Hypot(f.ncoCos, f.ncoSin)
			if magnitude > 0 {
				f.ncoCos, f.ncoSin = f.ncoCos/magnitude, f.ncoSin/magnitude
			}
		}
		historyIndex := int(f.inputSampleIndex % frontendHistory)
		f.iHistory[historyIndex], f.qHistory[historyIndex] = i, q
		f.inputSampleIndex++

		for float64(f.inputSampleIndex-1) >= f.nextOutputPosition+frontendFIRHalf {
			center := int64(math.Floor(f.nextOutputPosition))
			fraction := f.nextOutputPosition - float64(center)
			phase := min(max(int(math.Round(fraction*float64(frontendFIRPhases-1))), 0), frontendFIRPhases-1)
			cleanI, cleanQ := 0.0, 0.0
			for tap, coefficient := range f.kernel[phase] {
				sourceIndex := center + int64(tap-frontendFIRHalf)
				if sourceIndex < 0 || f.inputSampleIndex-sourceIndex > frontendHistory {
					continue
				}
				source := int(sourceIndex % frontendHistory)
				cleanI += f.iHistory[source] * coefficient
				cleanQ += f.qHistory[source] * coefficient
			}
			cross := cleanQ*f.previousI - cleanI*f.previousQ
			dot := cleanI*f.previousI + cleanQ*f.previousQ
			instantHz := math.Atan2(cross, dot) * outputRate / (2 * math.Pi)
			f.previousI, f.previousQ = cleanI, cleanQ
			errorSum += instantHz
			value := int(math.Round(instantHz / 2500 * 20000))
			result = append(result, int16(min(max(value, -32768), 32767)))
			f.nextOutputPosition += inputPerOutput
		}
	}
	if len(result) == 0 {
		return result, 0
	}
	return result, float32(errorSum / float64(len(result)))
}

func (f *frontend) setCorrection(hz float64) {
	f.mu.Lock()
	f.correctionHz = hz
	f.mu.Unlock()
}
