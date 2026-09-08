package sdr

import (
	"math"
	"testing"
)

func TestIQStats(t *testing.T) {
	rms, peak, invalid := iqStats([]float32{1, 0, 0, 1, float32(math.NaN()), 0})
	if math.Abs(rms-math.Sqrt(2.0/3.0)) > 1e-6 || peak != 1 || invalid != 1 {
		t.Fatalf("rms=%v peak=%v invalid=%d", rms, peak, invalid)
	}
}

func TestSquelchClosesAndReopensAudio(t *testing.T) {
	receiver := NewReceiver(Config{SampleRate: 2_048_000, FFTSize: 4096})
	receiver.SetSquelch(true, -100, 0, 20)

	closed := make([]float32, 1_200)
	for index := range closed {
		closed[index] = 1
	}
	receiver.mu.Lock()
	receiver.applySquelchLocked(closed, -110)
	receiver.mu.Unlock()
	if closed[len(closed)-1] != 0 || receiver.stats.SquelchOpen {
		t.Fatalf("squelch did not close: last=%v open=%v", closed[len(closed)-1], receiver.stats.SquelchOpen)
	}

	opened := make([]float32, 1_200)
	for index := range opened {
		opened[index] = 1
	}
	receiver.mu.Lock()
	receiver.applySquelchLocked(opened, -90)
	receiver.mu.Unlock()
	if opened[len(opened)-1] < .9 || !receiver.stats.SquelchOpen {
		t.Fatalf("squelch did not reopen: last=%v open=%v", opened[len(opened)-1], receiver.stats.SquelchOpen)
	}
}

func TestSignalLevelUsesTunedPassband(t *testing.T) {
	receiver := NewReceiver(Config{FrequencyHz: 100_000_000, SampleRate: 2_048_000, FFTSize: 4096})
	receiver.stats.FFTBlocks = 1
	receiver.spectrum[2048] = -82
	if got := receiver.signalLevelLocked("NFM", 100_000_000, 100_000_000, 12_500); got != -82 {
		t.Fatalf("signal level = %v, want -82", got)
	}
}

func TestDMRPublishesRFLevelForSMeter(t *testing.T) {
	receiver := NewReceiver(Config{FrequencyHz: 100_000_000, SampleRate: 2_048_000, FFTSize: 4096})
	receiver.mu.Lock()
	receiver.demodMode = "DMR BETA"
	receiver.tunedHz = 100_000_000
	receiver.demodBandwidthHz = 12_500
	receiver.stats.FFTBlocks = 1
	for index := range receiver.spectrum {
		receiver.spectrum[index] = -83
	}
	receiver.mu.Unlock()

	receiver.processAudio(make([]float32, 128))
	receiver.mu.RLock()
	level := receiver.stats.SignalDBm
	receiver.mu.RUnlock()
	if level != -83 {
		t.Fatalf("DMR SignalDBm=%v, want -83", level)
	}
}
