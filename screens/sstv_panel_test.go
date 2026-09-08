package screens

import "testing"

func TestSSTVUsesOneAutomaticAndThreeManualPreviewSelectors(t *testing.T) {
	screen := &MainScreen{sstvAutomatic: true, sstvMode: "R36", sstvCandidateModes: [4]string{"R36", "R72", "M1", "S1"}}
	panel := NewSSTVPanel(screen)
	if len(panel.candidateModes) != 3 {
		t.Fatalf("manual selector count = %d, want 3", len(panel.candidateModes))
	}
	for i, dropdown := range panel.candidateModes {
		if got, want := dropdown.SelectedText(), screen.sstvCandidateModes[i]; got != want {
			t.Fatalf("RX%d mode = %q, want %q", i+2, got, want)
		}
		if dropdown.Bounds().Y != toolY+34 {
			t.Fatalf("RX%d selector is not above its preview", i+2)
		}
	}
}
