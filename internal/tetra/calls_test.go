package tetra

import (
	"testing"
	"time"
)

func TestCallLifecycleAndUsageMapping(t *testing.T) {
	d := &Decoder{calls: make(map[uint16]Call)}
	now := time.Now()
	address := resourceAddress{SSI: 6009004, HasUsageMarker: true, UsageMarker: 17}
	d.updateCallLocked(cmceInfo{Kind: "D-SETUP", Code: 7, CallID: 1818}, address, 1, now)
	call := d.calls[1818]
	if !call.Active || call.Slot != 2 || call.UsageMarker != 17 || d.usageCall[17] != 1818 {
		t.Fatalf("unexpected active call: %#v mapping=%d", call, d.usageCall[17])
	}
	d.activeCallID, d.activeAudioSlot = 1818, 2
	d.updateCallLocked(cmceInfo{Kind: "D-RELEASE", Code: 6, CallID: 1818}, address, 1, now.Add(time.Second))
	call = d.calls[1818]
	if call.Active || d.activeCallID != 0 || d.activeAudioSlot != 0 {
		t.Fatalf("released call remains selected: %#v", call)
	}
}
