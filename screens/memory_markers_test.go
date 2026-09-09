package screens

import "testing"

func TestVisibleMemoryMarkersIgnoreTableFilters(t *testing.T) {
	screen := &MainScreen{centerFrequencyHz: 150_000_000, spanHz: 2_000_000}
	panel := &MemoryPanel{
		screen:        screen,
		selectedGroup: "AÉREA",
		onlyActive:    true,
		memories: []MemoryEntry{
			{Name: "A", Group: "AÉREA", FrequencyHz: 149_500_000, ScanEnabled: true},
			{Name: "B", Group: "MARINA", FrequencyHz: 150_500_000, ScanEnabled: false},
			{Name: "FUERA", Group: "AÉREA", FrequencyHz: 153_000_000, ScanEnabled: true},
		},
	}

	indices := panel.visibleMarkerIndices()
	if len(indices) != 2 || indices[0] != 0 || indices[1] != 1 {
		t.Fatalf("marcadores visibles = %v; se esperaban todas las memorias dentro del FFT", indices)
	}
}
