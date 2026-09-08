package screens

import (
	"strings"
	"testing"

	"go-zero/internal/rtl433"
)

func TestRTL433ClipboardUsesCompleteRawFrame(t *testing.T) {
	raw := `{"model":"Nexus-TH","id":199,"temperature_C":27.4,"humidity":55}`
	got := rtl433ClipboardText(rtl433.Event{Model: "Nexus-TH", Raw: raw})
	if got != raw {
		t.Fatalf("clipboard text was altered: %q", got)
	}
}

func TestRTL433ClipboardHasJSONFallback(t *testing.T) {
	got := rtl433ClipboardText(rtl433.Event{Model: "Nexus-TH", ID: "199"})
	if !strings.Contains(got, `"model":"Nexus-TH"`) || !strings.Contains(got, `"id":"199"`) {
		t.Fatalf("fallback is not a complete JSON event: %q", got)
	}
}
