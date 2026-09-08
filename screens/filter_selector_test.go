package screens

import "testing"

func TestAMFilterCatalog(t *testing.T) {
	presets := filterCatalog["AM"]
	expected := []int{10000, 6000, 4000, 7500}
	if len(presets) != 4 {
		t.Fatalf("expected four AM filters, got %d", len(presets))
	}
	for index, bandwidth := range expected {
		if presets[index].BandwidthHz != bandwidth {
			t.Fatalf("filter %d: got %d, want %d", index, presets[index].BandwidthHz, bandwidth)
		}
	}
}

func TestCustomAMFilterQuantization(t *testing.T) {
	selector := NewFilterSelector(nil)
	selector.mode = "AM"
	selector.setCustomNormalized(.5)
	bandwidth := filterCatalog["AM"][3].BandwidthHz
	if bandwidth < 2000 || bandwidth > 15000 || bandwidth%250 != 0 {
		t.Fatalf("invalid custom AM bandwidth: %d", bandwidth)
	}
	filterCatalog["AM"][3].BandwidthHz = 7500
}
