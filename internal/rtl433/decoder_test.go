package rtl433

import (
	"math"
	"testing"
)

func TestParseEventPreservesOriginalFields(t *testing.T) {
	line := []byte(`{"time":"@0.017384s","protocol":59,"type":"TPMS","model":"Steelmate","id":"0x2980","pressure_kPa":43.75,"temperature_C":23,"battery_mV":2990,"mod":"FSK","freq1":433.923,"rssi":1.167,"snr":6.408}`)
	event, ok := ParseEvent(line)
	if !ok || event.Protocol != 59 || event.Model != "Steelmate" || event.ID != "0x2980" {
		t.Fatalf("unexpected event: %+v, ok=%v", event, ok)
	}
	if math.Abs(event.FreqMHz-433.923) > .000001 || event.Summary == "" || event.Raw != string(line) {
		t.Fatalf("event lost decoder data: %+v", event)
	}
}

func TestFrontendProducesCU8AtOutputRate(t *testing.T) {
	f := newFrontend(2_048_000, 250_000, 250_000)
	iq := make([]float32, 4096*2)
	for n := 0; n < 4096; n++ {
		phase := 2 * math.Pi * 250_000 * float64(n) / 2_048_000
		iq[n*2], iq[n*2+1] = float32(math.Cos(phase))*.1, float32(math.Sin(phase))*.1
	}
	out := f.process(iq)
	if len(out) != 4096/8*2 {
		t.Fatalf("CU8 bytes=%d, want %d", len(out), 4096/8*2)
	}
}

func TestFrontendSelectableBandwidthRates(t *testing.T) {
	cases := []struct{ width, rate, decimation int }{{250_000, 256_000, 8}, {500_000, 512_000, 4}, {1_000_000, 1_024_000, 2}, {2_000_000, 2_048_000, 1}}
	for _, test := range cases {
		front := newFrontend(2_048_000, 0, test.width)
		if front.outputRate != test.rate || front.decimation != test.decimation {
			t.Errorf("width %d: rate=%d decimation=%d", test.width, front.outputRate, front.decimation)
		}
	}
}

func TestWideCoverageIsSplitIntoContiguousNarrowChannels(t *testing.T) {
	for _, test := range []struct {
		width, count int
	}{{1_000_000, 4}, {2_000_000, 8}} {
		centers := multichannelCenters(433_920_000, test.width)
		if len(centers) != test.count {
			t.Fatalf("width %d: channels=%d, want %d", test.width, len(centers), test.count)
		}
		for index := 1; index < len(centers); index++ {
			if centers[index]-centers[index-1] != 250_000 {
				t.Fatalf("width %d: channel gap is not 250 kHz: %v", test.width, centers)
			}
		}
		coveredLow := centers[0] - 125_000
		coveredHigh := centers[len(centers)-1] + 125_000
		if coveredHigh-coveredLow != int64(test.width) {
			t.Fatalf("width %d: coverage %d..%d", test.width, coveredLow, coveredHigh)
		}
	}
}
